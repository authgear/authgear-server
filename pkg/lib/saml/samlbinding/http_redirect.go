package samlbinding

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"

	"github.com/beevik/etree"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
)

type SAMLBindingHTTPRedirectParseResult struct {
	SAMLRequest    string
	SAMLRequestXML string
	RelayState     string
	SigAlg         string
	Signature      string
}

var _ SAMLBindingParseResult = &SAMLBindingHTTPRedirectParseResult{}

func (*SAMLBindingHTTPRedirectParseResult) samlBindingParseResult() {}

func SAMLBindingHTTPRedirectParse(r *http.Request) (
	result *SAMLBindingHTTPRedirectParseResult,
	err error,
) {
	result = &SAMLBindingHTTPRedirectParseResult{}
	relayState := r.URL.Query().Get("RelayState")
	result.RelayState = relayState
	signature := r.URL.Query().Get("Signature")
	sigAlg := r.URL.Query().Get("SigAlg")
	samlRequest := r.URL.Query().Get("SAMLRequest")
	result.SAMLRequest = samlRequest
	if samlRequest == "" {
		return nil, ErrNoRequest
	}
	compressedRequest, err := base64.StdEncoding.DecodeString(samlRequest)
	if err != nil {
		return result, &samlprotocol.ParseRequestFailedError{
			Reason: "base64 decode failed",
			Cause:  err,
		}
	}
	requestBuffer, err := io.ReadAll(newSaferFlateReader(bytes.NewReader(compressedRequest)))
	if err != nil {
		return result, &samlprotocol.ParseRequestFailedError{
			Reason: "decompress failed",
			Cause:  err,
		}
	}

	result.SAMLRequestXML = string(requestBuffer)
	result.Signature = signature
	result.SigAlg = sigAlg

	return result, nil
}

type SAMLBindingHTTPRedirectWriter struct {
	Signer SAMLRedirectBindingSigner
}

func (s *SAMLBindingHTTPRedirectWriter) Write(
	rw http.ResponseWriter,
	r *http.Request,
	callbackURL string,
	response samlprotocol.Respondable,
	relayState string) error {

	responseEl := response.Element()

	// https://docs.oasis-open.org/security/saml/v2.0/saml-bindings-2.0-os.pdf
	// 3.4.4.1 DEFLATE Encoding
	// Any signature on the SAML protocol message, including the <ds:Signature> XML element itself,
	// MUST be removed.
	if sigEl := responseEl.FindElement("./Signature"); sigEl != nil {
		responseEl.RemoveChild(sigEl)
	}

	doc := etree.NewDocument()
	doc.SetRoot(responseEl)
	responseBuf, err := doc.WriteToBytes()
	if err != nil {
		return err
	}

	compressedResponseBuffer := &bytes.Buffer{}
	writer, err := flate.NewWriter(compressedResponseBuffer, 9)
	if err != nil {
		return err
	}
	_, err = writer.Write(responseBuf)
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	encodedResponse := base64.StdEncoding.EncodeToString(compressedResponseBuffer.Bytes())

	redirectURL, err := url.Parse(callbackURL)
	if err != nil {
		return err
	}

	q, err := s.Signer.ConstructSignedQueryParameters(encodedResponse, relayState)
	if err != nil {
		return err
	}

	redirectURLQuery := redirectURL.Query()
	for key, values := range q {
		for _, v := range values {
			redirectURLQuery.Add(key, v)
		}
	}

	redirectURL.RawQuery = redirectURLQuery.Encode()

	http.Redirect(rw, r, redirectURL.String(), http.StatusFound)
	return nil
}
