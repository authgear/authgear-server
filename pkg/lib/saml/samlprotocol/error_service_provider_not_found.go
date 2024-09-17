package samlprotocol

import (
	"fmt"

	"github.com/beevik/etree"
)

type serviceProviderNotFoundError struct {
}

var _ SAMLErrorCodeError = &serviceProviderNotFoundError{}

func (s *serviceProviderNotFoundError) Error() string {
	return fmt.Sprintf("saml error(%s): service provider not found",
		s.ErrorCode())
}
func (s *serviceProviderNotFoundError) ErrorCode() SAMLErrorCode {
	return SAMLErrorCodeServiceProviderNotFound
}
func (s *serviceProviderNotFoundError) GetDetailElements() []*etree.Element {
	codeEl := etree.NewElement("ErrorCode")
	codeEl.SetText(string(s.ErrorCode()))
	els := []*etree.Element{
		codeEl,
	}
	return els
}

var ErrServiceProviderNotFound = &serviceProviderNotFoundError{}
