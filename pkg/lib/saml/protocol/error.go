package protocol

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/saml"
)

type SAMLProtocolError struct {
	Response   *saml.Response
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
