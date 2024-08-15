package saml

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/saml"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlsession"
)

type HandlerSAMLService interface {
	IdpMetadata(serviceProviderId string) (*saml.Metadata, error)
	ValidateAuthnRequest(serviceProviderId string, authnRequest *saml.AuthnRequest) error
}

type SAMLSessionService interface {
	Save(entry *samlsession.SAMLSession) (err error)
	Get(entryID string) (*samlsession.SAMLSession, error)
	Delete(entryID string) error
}

type SAMLUIURLBuilder interface {
	BuildAuthenticationURL(s *samlsession.SAMLSession) (*url.URL, error)
}
