package samlprotocol

import (
	"fmt"

	"github.com/beevik/etree"
)

type UnsupportedAttributeTypeError struct {
	AttributeName      string
	UserProfilePointer string
}

var _ SAMLErrorCodeError = &UnsupportedAttributeTypeError{}

func (s *UnsupportedAttributeTypeError) Error() string {
	return fmt.Sprintf("saml error(%s): unsupported_attribute_type: name:%v user_profile_pointer:%v",
		s.ErrorCode(),
		s.AttributeName,
		s.UserProfilePointer)
}
func (s *UnsupportedAttributeTypeError) ErrorCode() SAMLErrorCode {
	return SAMLUnsupportedAttributeType
}
func (s *UnsupportedAttributeTypeError) GetDetailElements() []*etree.Element {
	codeEl := etree.NewElement("ErrorCode")
	codeEl.SetText(string(s.ErrorCode()))
	els := []*etree.Element{
		codeEl,
	}
	if s.AttributeName != "" {
		el := etree.NewElement("AttributeName")
		el.SetText(s.AttributeName)
		els = append(els, el)
	}
	if s.UserProfilePointer != "" {
		el := etree.NewElement("UserProfilePointer")
		el.SetText(s.UserProfilePointer)
		els = append(els, el)
	}

	return els
}
