package saml

import (
	"context"
	"net/http"
	"net/url"

	"github.com/beevik/etree"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/saml"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlsession"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlslosession"
	"github.com/authgear/authgear-server/pkg/lib/session"
)

type HandlerSAMLService interface {
	IdpEntityID() string
	IdpMetadata(serviceProviderId string) (*samlprotocol.Metadata, error)
	ValidateAuthnRequest(serviceProviderId string, authnRequest *samlprotocol.AuthnRequest) error
	IssueLoginSuccessResponse(
		ctx context.Context,
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
		element *saml.SAMLElementSigned,
		sigAlg string,
		relayState string,
		signature string) error
	IssueLogoutResponse(
		callbackURL string,
		serviceProviderId string,
		inResponseToLogoutRequest *samlprotocol.LogoutRequest,
		isPartialLogout bool,
	) (*samlprotocol.LogoutResponse, error)
	IssueLogoutRequest(
		sp *config.SAMLServiceProviderConfig,
		sloSession *samlslosession.SAMLSLOSession,
	) (*samlprotocol.LogoutRequest, error)
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

type BindingHTTPPostWriter interface {
	WriteResponse(
		rw http.ResponseWriter,
		r *http.Request,
		callbackURL string,
		responseElement *etree.Element,
		relayState string) error
	WriteRequest(
		rw http.ResponseWriter,
		r *http.Request,
		callbackURL string,
		requestElement *etree.Element,
		relayState string) error
}

type BindingHTTPRedirectWriter interface {
	WriteResponse(
		rw http.ResponseWriter,
		r *http.Request,
		callbackURL string,
		responseElement *etree.Element,
		relayState string) error
	WriteRequest(
		rw http.ResponseWriter,
		r *http.Request,
		callbackURL string,
		requestElement *etree.Element,
		relayState string) error
}

type SessionManager interface {
	Get(id string) (session.ListableSession, error)
	Logout(session.SessionBase, http.ResponseWriter) ([]session.ListableSession, error)
}

type SAMLSLOSessionService interface {
	Get(sessionID string) (entry *samlslosession.SAMLSLOSession, err error)
	Save(session *samlslosession.SAMLSLOSession) (err error)
}

type SAMLSLOService interface {
	SendSLORequest(
		rw http.ResponseWriter,
		r *http.Request,
		sloSession *samlslosession.SAMLSLOSession,
		sp *config.SAMLServiceProviderConfig,
	) error
}
