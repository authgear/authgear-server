package samlbinding

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlerror"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
)

type SAMLBindingHTTPRedirectParser struct{}

type SAMLBindingHTTPRedirectParseResult struct {
	AuthnRequest *samlprotocol.AuthnRequest
	RelayState   string
	SignedValue  string
	SigAlg       string
	Signature    string
}

var _ SAMLBindingParseResult = &SAMLBindingHTTPRedirectParseResult{}

func (*SAMLBindingHTTPRedirectParseResult) SAMLBindingParseResult() {}

func (*SAMLBindingHTTPRedirectParser) Parse(r *http.Request) (
	result *SAMLBindingHTTPRedirectParseResult,
	err error,
) {
	result = &SAMLBindingHTTPRedirectParseResult{}
	relayState := r.URL.Query().Get("RelayState")
	result.RelayState = relayState
	signature := r.URL.Query().Get("Signature")
	sigAlg := r.URL.Query().Get("SigAlg")
	samlRequest := r.URL.Query().Get("SAMLRequest")
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

	request, err := samlprotocol.ParseAuthnRequest(requestBuffer)
	if err != nil {
		return result, &samlerror.ParseRequestFailedError{
			Reason: "malformed AuthnRequest",
			Cause:  err,
		}
	}

	result.AuthnRequest = request
	result.Signature = signature
	result.SigAlg = sigAlg

	// https://docs.oasis-open.org/security/saml/v2.0/saml-bindings-2.0-os.pdf 3.4.4.1
	signedValues := []string{}
	signedValues = append(signedValues, fmt.Sprintf("SAMLRequest=%s", url.QueryEscape(samlRequest)))
	if relayState != "" {
		signedValues = append(signedValues, fmt.Sprintf("RelayState=%s", url.QueryEscape(relayState)))
	}
	if sigAlg != "" {
		signedValues = append(signedValues, fmt.Sprintf("SigAlg=%s", url.QueryEscape(sigAlg)))

	}

	result.SignedValue = strings.Join(signedValues, "&")

	return result, nil
}
