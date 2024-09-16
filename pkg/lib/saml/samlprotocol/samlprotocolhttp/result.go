package samlprotocolhttp

import (
	"fmt"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
)

type SAMLResultSigner interface {
	ConstructSignedQueryParameters(
		samlResponse string,
		relayState string,
	) (url.Values, error)
}

type SAMLResult interface {
	GetResponse() samlprotocol.Respondable
}

type SAMLSuccessResult struct {
	Response samlprotocol.Respondable
}

var _ SAMLResult = &SAMLSuccessResult{}

func (r *SAMLSuccessResult) GetResponse() samlprotocol.Respondable {
	return r.Response
}

type SAMLErrorResult struct {
	Response     samlprotocol.Respondable
	Cause        error
	IsUnexpected bool
}

var _ SAMLResult = &SAMLErrorResult{}

func (r *SAMLErrorResult) GetResponse() samlprotocol.Respondable {
	return r.Response
}

var _ error = &SAMLErrorResult{}

func (s *SAMLErrorResult) Error() string {
	return fmt.Sprintf("saml error response: %v", s.Cause)
}

func (s *SAMLErrorResult) Unwrap() error {
	return s.Cause
}

func (s *SAMLErrorResult) IsInternalError() bool {
	return s.IsUnexpected
}

func NewExpectedSAMLErrorResult(cause error, response samlprotocol.Respondable) *SAMLErrorResult {
	return &SAMLErrorResult{
		Response:     response,
		Cause:        cause,
		IsUnexpected: false,
	}
}

func NewUnexpectedSAMLErrorResult(cause error, response samlprotocol.Respondable) *SAMLErrorResult {
	return &SAMLErrorResult{
		Response:     response,
		Cause:        cause,
		IsUnexpected: true,
	}
}
