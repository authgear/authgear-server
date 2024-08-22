package samlerror

import (
	"fmt"

	"github.com/beevik/etree"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
)

type MissingNameIDError struct {
	ExpectedNameIDFormat   string
	NameIDAttributePointer string
}

var _ samlprotocol.SAMLErrorCodeError = &MissingNameIDError{}

func (s *MissingNameIDError) Error() string {
	return fmt.Sprintf("saml error(%s): expected_nameid_format:%v",
		s.ErrorCode(),
		s.ExpectedNameIDFormat)
}
func (s *MissingNameIDError) ErrorCode() samlprotocol.SAMLErrorCode {
	return samlprotocol.SAMLErrorCodeMissingNameID
}
func (s *MissingNameIDError) GetDetailElements() []*etree.Element {
	codeEl := etree.NewElement("ErrorCode")
	codeEl.SetText(string(s.ErrorCode()))
	els := []*etree.Element{
		codeEl,
	}
	if s.ExpectedNameIDFormat != "" {
		el := etree.NewElement("ExpectedNameIDFormat")
		el.SetText(s.ExpectedNameIDFormat)
		els = append(els, el)
	}
	if s.NameIDAttributePointer != "" {
		el := etree.NewElement("NameIDAttributePointer")
		el.SetText(s.NameIDAttributePointer)
		els = append(els, el)
	}

	return els
}
