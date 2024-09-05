package samlerror

import (
	"fmt"

	"github.com/beevik/etree"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
)

type InvalidSignatureError struct {
	Cause error
}

var _ samlprotocol.SAMLErrorCodeError = &InvalidSignatureError{}

func (s *InvalidSignatureError) Error() string {
	return fmt.Sprintf("saml error(%s): cause:%v",
		s.ErrorCode(),
		s.Cause)
}
func (s *InvalidSignatureError) ErrorCode() samlprotocol.SAMLErrorCode {
	return samlprotocol.SAMLErrorCodeInvalidSignature
}
func (s *InvalidSignatureError) GetDetailElements() []*etree.Element {
	codeEl := etree.NewElement("ErrorCode")
	codeEl.SetText(string(s.ErrorCode()))
	els := []*etree.Element{
		codeEl,
	}
	return els
}

func (s *InvalidSignatureError) Unwrap() error {
	return s.Cause
}
