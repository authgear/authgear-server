package saml

import (
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlsession"
)

type HandlerSAMLService interface {
	IdpMetadata(serviceProviderId string) (*samlprotocol.Metadata, error)
	ValidateAuthnRequest(serviceProviderId string, authnRequest *samlprotocol.AuthnRequest) error
	IssueSuccessResponse(
		serviceProviderId string,
		authenticatedUserId string,
		inResponseToAuthnRequest *samlprotocol.AuthnRequest,
	) (*samlprotocol.Response, error)
}

type SAMLSessionService interface {
	Save(entry *samlsession.SAMLSession) (err error)
	Get(entryID string) (*samlsession.SAMLSession, error)
	Delete(entryID string) error
}

type SAMLUIService interface {
	ResolveUIInfo(entry *samlsession.SAMLSessionEntry) (*samlsession.SAMLUIInfo, error)
	BuildAuthenticationURL(s *samlsession.SAMLSession) (*url.URL, error)
}

type SAMLAuthenticationInfoResolver interface {
	GetAuthenticationInfoID(req *http.Request) (string, bool)
}

type SAMLAuthenticationInfoService interface {
	Get(entryID string) (*authenticationinfo.Entry, error)
	Delete(entryID string) error
}
