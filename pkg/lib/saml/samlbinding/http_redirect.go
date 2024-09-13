package samlbinding

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlerror"
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
		return result, &samlerror.ParseRequestFailedError{
			Reason: "base64 decode failed",
			Cause:  err,
		}
	}
	requestBuffer, err := io.ReadAll(newSaferFlateReader(bytes.NewReader(compressedRequest)))
	if err != nil {
		return result, &samlerror.ParseRequestFailedError{
			Reason: "decompress failed",
			Cause:  err,
		}
	}

	result.SAMLRequestXML = string(requestBuffer)
	result.Signature = signature
	result.SigAlg = sigAlg

	return result, nil
}
