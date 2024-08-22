package samlerror

import (
	"fmt"

	"github.com/beevik/etree"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
)

type serviceProviderNotFoundError struct {
}

var _ samlprotocol.SAMLErrorCodeError = &serviceProviderNotFoundError{}

func (s *serviceProviderNotFoundError) Error() string {
	return fmt.Sprintf("saml error(%s): service provider not found",
		s.ErrorCode())
}
func (s *serviceProviderNotFoundError) ErrorCode() samlprotocol.SAMLErrorCode {
	return samlprotocol.SAMLErrorCodeServiceProviderNotFound
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
