package samlprotocol

import (
	"fmt"

	"github.com/beevik/etree"
)

type InvalidRequestError struct {
	Field    string
	Actual   string
	Expected []string
	Reason   string
}

var _ SAMLErrorCodeError = &InvalidRequestError{}

func (s *InvalidRequestError) Error() string {
	return fmt.Sprintf("saml error(%s): field:%v expected:%v actual:%v reason:%v",
		s.ErrorCode(),
		s.Field, s.Expected, s.Actual, s.Reason)
}
func (s *InvalidRequestError) ErrorCode() SAMLErrorCode {
	return SAMLErrorCodeInvalidRequest
}
func (s *InvalidRequestError) GetDetailElements() []*etree.Element {
	codeEl := etree.NewElement("ErrorCode")
	codeEl.SetText(string(s.ErrorCode()))
	els := []*etree.Element{
		codeEl,
	}
	if s.Field != "" {
		fieldEl := etree.NewElement("Field")
		fieldEl.SetText(s.Field)
		els = append(els, fieldEl)
	}
	if s.Actual != "" {
		actualEl := etree.NewElement("Actual")
		actualEl.SetText(s.Actual)
		els = append(els, actualEl)
	}
	for _, expected := range s.Expected {
		expectedEl := etree.NewElement("Expected")
		expectedEl.SetText(expected)
		els = append(els, expectedEl)
	}

	if s.Reason != "" {
		reasonEl := etree.NewElement("Reason")
		reasonEl.SetText(s.Reason)
		els = append(els, reasonEl)
	}

	return els
}
