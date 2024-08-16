package samlbinding

import (
	"encoding/base64"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlerror"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
)

type SAMLBindingHTTPPostParser struct{}

type SAMLBindingHTTPPostParseResult struct {
	AuthnRequest *samlprotocol.AuthnRequest
	RelayState   string
}

var _ SAMLBindingParseResult = &SAMLBindingHTTPPostParseResult{}

func (*SAMLBindingHTTPPostParseResult) SAMLBindingParseResult() {}

func (*SAMLBindingHTTPPostParser) Parse(r *http.Request) (
	result *SAMLBindingHTTPPostParseResult,
	err error,
) {
	result = &SAMLBindingHTTPPostParseResult{}
	if err := r.ParseForm(); err != nil {
		return result, &samlerror.ParseRequestFailedError{
			Reason: "failed to parse request body as a application/x-www-form-urlencoded form",
			Cause:  err,
		}
	}
	relayState := r.PostForm.Get("RelayState")
	result.RelayState = relayState

	requestBuffer, err := base64.StdEncoding.DecodeString(r.PostForm.Get("SAMLRequest"))
	if err != nil {
		return result, &samlerror.ParseRequestFailedError{
			Reason: "base64 decode failed",
			Cause:  err,
		}
	}

	authnRequest, err := samlprotocol.ParseAuthnRequest(requestBuffer)
	if err != nil {
		return result, &samlerror.ParseRequestFailedError{
			Reason: "malformed AuthnRequest",
			Cause:  err,
		}
	}
	result.AuthnRequest = authnRequest

	return result, nil
}
