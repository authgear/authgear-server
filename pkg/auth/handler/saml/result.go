package saml

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
)

type SAMLErrorResult struct {
	Response samlprotocol.Respondable
	Cause    error
}

var _ error = &SAMLErrorResult{}

func (s *SAMLErrorResult) Error() string {
	return fmt.Sprintf("saml error response: %v", s.Cause)
}

func (s *SAMLErrorResult) Unwrap() error {
	return s.Cause
}

func NewSAMLErrorResult(cause error, response samlprotocol.Respondable) *SAMLErrorResult {
	return &SAMLErrorResult{
		Response: response,
		Cause:    cause,
	}
}
