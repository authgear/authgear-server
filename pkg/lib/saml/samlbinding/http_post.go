package samlbinding

import (
	"encoding/base64"
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
)

type SAMLBindingHTTPPostParser struct{}

type SAMLBindingHTTPPostParseResult struct {
	AuthnRequest *samlprotocol.AuthnRequest
	RelayState   string
}

var _ SAMLBindingParseResult = &SAMLBindingHTTPPostParseResult{}

func (*SAMLBindingHTTPPostParseResult) SAMLBindingParseResult() {}

func (*SAMLBindingHTTPPostParser) Parse(now time.Time, r *http.Request) (
	result *SAMLBindingHTTPPostParseResult,
	err error,
) {
	result = &SAMLBindingHTTPPostParseResult{}
	if err := r.ParseForm(); err != nil {
		return result, &samlprotocol.SAMLProtocolError{
			Response: samlprotocol.NewRequestDeniedErrorResponse(now, "failed to parse request body"),
			Cause:    err,
		}
	}
	relayState := r.PostForm.Get("RelayState")
	result.RelayState = relayState

	requestBuffer, err := base64.StdEncoding.DecodeString(r.PostForm.Get("SAMLRequest"))
	if err != nil {
		return result, &samlprotocol.SAMLProtocolError{
			Response:   samlprotocol.NewRequestDeniedErrorResponse(now, "failed to decode SAMLRequest"),
			RelayState: relayState,
			Cause:      err,
		}
	}

	authnRequest, err := samlprotocol.ParseAuthnRequest(requestBuffer)
	if err != nil {
		return result, &samlprotocol.SAMLProtocolError{
			Response:   samlprotocol.NewRequestDeniedErrorResponse(now, "failed to parse SAMLRequest"),
			RelayState: relayState,
			Cause:      err,
		}
	}
	result.AuthnRequest = authnRequest

	return result, nil
}
