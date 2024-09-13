package saml

import (
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlsession"
)

type HandlerSAMLService interface {
	IdpEntityID() string
	IdpMetadata(serviceProviderId string) (*samlprotocol.Metadata, error)
	ValidateAuthnRequest(serviceProviderId string, authnRequest *samlprotocol.AuthnRequest) error
	IssueSuccessResponse(
		callbackURL string,
		serviceProviderId string,
		authInfo authenticationinfo.T,
		inResponseToAuthnRequest *samlprotocol.AuthnRequest,
	) (*samlprotocol.Response, error)
	VerifyEmbeddedSignature(
		sp *config.SAMLServiceProviderConfig,
		samlRequestXML string) error
	VerifyExternalSignature(
		sp *config.SAMLServiceProviderConfig,
		samlRequest string,
		sigAlg string,
		relayState string,
		signature string) error
	IssueLogoutResponse(
		callbackURL string,
		serviceProviderId string,
		inResponseToLogoutRequest *samlprotocol.LogoutRequest,
	) (*samlprotocol.LogoutResponse, error)
}

type SAMLSessionService interface {
	Save(entry *samlsession.SAMLSession) (err error)
	Get(entryID string) (*samlsession.SAMLSession, error)
	Delete(entryID string) error
}

type SAMLUIService interface {
	ResolveUIInfo(
		sp *config.SAMLServiceProviderConfig,
		entry *samlsession.SAMLSessionEntry,
	) (info *samlsession.SAMLUIInfo, showUI bool, err error)
	BuildAuthenticationURL(s *samlsession.SAMLSession) (*url.URL, error)
}

type SAMLAuthenticationInfoResolver interface {
	GetAuthenticationInfoID(req *http.Request) (string, bool)
}

type SAMLAuthenticationInfoService interface {
	Get(entryID string) (*authenticationinfo.Entry, error)
	Delete(entryID string) error
}

type SAMLUserFacade interface {
	GetUserIDsByLoginHint(hint *oauth.LoginHint) ([]string, error)
}
