package samlprotocol

import (
	"fmt"

	"github.com/beevik/etree"
)

// This error represents an error that can be sent through SAML bindings
type SAMLErrorResponse struct {
	Response   *Response
	RelayState string
	Cause      error
}

func (s *SAMLErrorResponse) Error() string {
	return fmt.Sprintf("saml error response: %v", s.Cause)
}

func (s *SAMLErrorResponse) Unwrap() error {
	return s.Cause
}

var _ error = &SAMLErrorResponse{}

type SAMLErrorCode string

const (
	SAMLErrorCodeServiceProviderNotFound SAMLErrorCode = "service_provider_not_found"
	SAMLErrorCodeInvalidRequest          SAMLErrorCode = "invalid_request"
	SAMLErrorCodeParseRequestFailed      SAMLErrorCode = "parse_request_failed"
)

// This error can be thrown in any code related to SAML, mainly in saml.Service
type SAMLErrorCodeError interface {
	error
	ErrorCode() SAMLErrorCode
	GetDetailElements() []*etree.Element
}
