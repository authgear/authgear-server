package samlprotocol

import (
	"fmt"

	"github.com/beevik/etree"
)

type ParseRequestFailedError struct {
	Reason string
	Cause  error
}

var _ SAMLErrorCodeError = &ParseRequestFailedError{}

func (s *ParseRequestFailedError) Error() string {
	return fmt.Sprintf("saml error(%s): %v",
		s.ErrorCode(),
		s.Cause)
}
func (s *ParseRequestFailedError) ErrorCode() SAMLErrorCode {
	return SAMLErrorCodeParseRequestFailed
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
