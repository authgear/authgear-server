package saml

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
)

type SAMLErrorResult struct {
	Response     samlprotocol.Respondable
	Cause        error
	IsUnexpected bool
}

var _ error = &SAMLErrorResult{}

func (s *SAMLErrorResult) Error() string {
	return fmt.Sprintf("saml error response: %v", s.Cause)
}

func (s *SAMLErrorResult) Unwrap() error {
	return s.Cause
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
