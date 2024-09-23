package samlbinding

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"

	"github.com/beevik/etree"

	"github.com/authgear/authgear-server/pkg/lib/saml"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
)

type SAMLBindingHTTPRedirectParseRequestResult struct {
	SAMLRequest    string
	SAMLRequestXML string
	RelayState     string
	SigAlg         string
	Signature      string
}

var _ SAMLBindingParseReqeustResult = &SAMLBindingHTTPRedirectParseRequestResult{}

func (*SAMLBindingHTTPRedirectParseRequestResult) samlBindingParseRequestResult() {}

func SAMLBindingHTTPRedirectParseRequest(r *http.Request) (
	result *SAMLBindingHTTPRedirectParseRequestResult,
	err error,
) {
	result = &SAMLBindingHTTPRedirectParseRequestResult{}
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

type SAMLBindingHTTPRedirectParseResponseResult struct {
	SAMLResponse    string
	SAMLResponseXML string
	RelayState      string
	SigAlg          string
	Signature       string
}

var _ SAMLBindingParseResponseResult = &SAMLBindingHTTPRedirectParseResponseResult{}

func (*SAMLBindingHTTPRedirectParseResponseResult) samlBindingParseResponseResult() {}

func SAMLBindingHTTPRedirectParseResponse(r *http.Request) (
	result *SAMLBindingHTTPRedirectParseResponseResult,
	err error,
) {
	result = &SAMLBindingHTTPRedirectParseResponseResult{}
	relayState := r.URL.Query().Get("RelayState")
	result.RelayState = relayState
	signature := r.URL.Query().Get("Signature")
	sigAlg := r.URL.Query().Get("SigAlg")
	samlResponse := r.URL.Query().Get("SAMLResponse")
	result.SAMLResponse = samlResponse
	if samlResponse == "" {
		return nil, ErrNoResponse
	}
	compressedResponse, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		return result, &samlprotocol.ParseRequestFailedError{
			Reason: "base64 decode failed",
			Cause:  err,
		}
	}
	responseBuffer, err := io.ReadAll(newSaferFlateReader(bytes.NewReader(compressedResponse)))
	if err != nil {
		return result, &samlprotocol.ParseRequestFailedError{
			Reason: "decompress failed",
			Cause:  err,
		}
	}

	result.SAMLResponseXML = string(responseBuffer)
	result.Signature = signature
	result.SigAlg = sigAlg

	return result, nil
}

type SAMLBindingHTTPRedirectWriter struct {
	Signer SAMLRedirectBindingSigner
}

type writeElements struct {
	SAMLResponse *etree.Element
	SAMLRequest  *etree.Element
}

func (s *SAMLBindingHTTPRedirectWriter) write(
	rw http.ResponseWriter,
	r *http.Request,
	callbackURL string,
	relayState string,
	elements *writeElements) error {
	var el *etree.Element
	if elements.SAMLRequest != nil {
		el = elements.SAMLRequest
	} else if elements.SAMLResponse != nil {
		el = elements.SAMLResponse
	} else {
		panic("no SAMLRequest or SAMLResponse given")
	}
	// https://docs.oasis-open.org/security/saml/v2.0/saml-bindings-2.0-os.pdf
	// 3.4.4.1 DEFLATE Encoding
	// Any signature on the SAML protocol message, including the <ds:Signature> XML element itself,
	// MUST be removed.
	if sigEl := el.FindElement("./Signature"); sigEl != nil {
		el.RemoveChild(sigEl)
	}

	doc := etree.NewDocument()
	doc.SetRoot(el)
	elBuf, err := doc.WriteToBytes()
	if err != nil {
		return err
	}

	compressedElBuffer := &bytes.Buffer{}
	writer, err := flate.NewWriter(compressedElBuffer, 9)
	if err != nil {
		return err
	}
	_, err = writer.Write(elBuf)
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	encodedEl := base64.StdEncoding.EncodeToString(compressedElBuffer.Bytes())

	redirectURL, err := url.Parse(callbackURL)
	if err != nil {
		return err
	}

	var elToSign *saml.SAMLElementToSign
	if elements.SAMLRequest != nil {
		elToSign = &saml.SAMLElementToSign{
			SAMLRequest: encodedEl,
		}
	} else if elements.SAMLResponse != nil {
		elToSign = &saml.SAMLElementToSign{
			SAMLResponse: encodedEl,
		}
	} else {
		panic("no SAMLRequest or SAMLResponse given")
	}

	q, err := s.Signer.ConstructSignedQueryParameters(relayState, elToSign)
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

func (s *SAMLBindingHTTPRedirectWriter) WriteResponse(
	rw http.ResponseWriter,
	r *http.Request,
	callbackURL string,
	responseEl *etree.Element,
	relayState string) error {
	return s.write(
		rw, r,
		callbackURL,
		relayState,
		&writeElements{
			SAMLResponse: responseEl,
		},
	)
}

func (s *SAMLBindingHTTPRedirectWriter) WriteRequest(
	rw http.ResponseWriter,
	r *http.Request,
	callbackURL string,
	requestEl *etree.Element,
	relayState string) error {
	return s.write(
		rw, r,
		callbackURL,
		relayState,
		&writeElements{
			SAMLRequest: requestEl,
		},
	)
}
