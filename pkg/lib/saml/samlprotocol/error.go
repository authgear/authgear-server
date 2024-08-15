package samlprotocol

import (
	"fmt"
)

type SAMLProtocolError struct {
	Response   *Response
	RelayState string
	Cause      error
}

func (s *SAMLProtocolError) Error() string {
	return fmt.Sprintf("saml error: %s", s.Cause.Error())
}

func (s *SAMLProtocolError) Unwrap() error {
	return s.Cause
}

var _ error = &SAMLProtocolError{}
