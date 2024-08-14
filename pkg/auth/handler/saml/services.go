package saml

import "github.com/authgear/authgear-server/pkg/lib/saml"

type HandlerSAMLService interface {
	IdpMetadata(serviceProviderId string) (*saml.Metadata, error)
	ValidateAuthnRequest(serviceProviderId string, authnRequest *saml.AuthnRequest) error
}
