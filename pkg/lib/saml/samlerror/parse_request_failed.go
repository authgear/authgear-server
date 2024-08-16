package samlerror

import (
	"fmt"

	"github.com/beevik/etree"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
)

type ParseRequestFailedError struct {
	Reason string
	Cause  error
}

var _ samlprotocol.SAMLErrorCodeError = &ParseRequestFailedError{}

func (s *ParseRequestFailedError) Error() string {
	return fmt.Sprintf("saml error(%s): %v",
		s.ErrorCode(),
		s.Cause)
}
func (s *ParseRequestFailedError) ErrorCode() samlprotocol.SAMLErrorCode {
	return samlprotocol.SAMLErrorCodeParseRequestFailed
}
func (s *ParseRequestFailedError) GetDetailElements() []*etree.Element {
	codeEl := etree.NewElement("ErrorCode")
	codeEl.SetText(string(s.ErrorCode()))
	els := []*etree.Element{
		codeEl,
	}
	if s.Reason != "" {
		reasonEl := etree.NewElement("Reason")
		reasonEl.SetText(s.Reason)
		els = append(els, reasonEl)
	}

	return els
}

func (s *ParseRequestFailedError) Unwrap() error {
	return s.Cause
}
