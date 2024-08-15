package samlsession

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/uiparam"
)

const (
	queryNameSAMLSessionID string = "x_saml_session_id"
)

type SAMLUIInfo struct {
	// SAMLServiceProviderID is id of the service provider
	SAMLServiceProviderID string
	// RedirectURI is the redirect_uri the UI should redirect to.
	// The redirect_uri in the URL has lower precedence.
	// The rationale for this is if the end-user bookmarked the
	// authorization URL in the browser, redirect to the app is
	// possible.
	RedirectURI string
	// Prompt is the resolved oidc prompt from ForceAuthn and IsPassive for AuthnRequest.
	Prompt []string
}

func (i *SAMLUIInfo) ToUIParam() uiparam.T {
	return uiparam.T{
		Prompt: i.Prompt,
	}
}

type UIServiceAuthUIEndpointsProvider interface {
	OAuthEntrypointURL() *url.URL
	SAMLLoginFinishURL(serviceProviderId string) *url.URL
}

type UIService struct {
	Endpoints UIServiceAuthUIEndpointsProvider
}

func (r *UIService) GetSAMLSessionID(req *http.Request, urlQuery string) (string, bool) {
	if q, err := url.ParseQuery(urlQuery); err == nil {
		id := q.Get(queryNameSAMLSessionID)
		if id != "" {
			return id, true
		}
	}

	id := req.URL.Query().Get(queryNameSAMLSessionID)
	if id != "" {
		return id, true
	}
	return "", false
}

func (r *UIService) RemoveSAMLSessionID(w http.ResponseWriter, req *http.Request) {
	// Remove from http.Request.URL
	urlQuery := req.URL.Query()
	urlQuery.Del(queryNameSAMLSessionID)
	reqURL := *req.URL
	reqURL.RawQuery = urlQuery.Encode()
	req.URL = &reqURL
}

func (r *UIService) ResolveUIInfo(entry *SAMLSessionEntry) (*SAMLUIInfo, error) {
	prompt := []string{}
	authnRequest := entry.AuthnRequest()
	switch {
	case authnRequest.GetIsPassive() == false && authnRequest.GetForceAuthn() == false:
		prompt = []string{"select_account"}
	case authnRequest.GetIsPassive() == false && authnRequest.GetForceAuthn() == true:
		prompt = []string{"login"}
	case authnRequest.GetIsPassive() == true && authnRequest.GetForceAuthn() == false:
		prompt = []string{"none"}
	default:
		return nil, fmt.Errorf("unsupported: IsPassive=true and ForceAuthn=true")
	}

	info := &SAMLUIInfo{
		SAMLServiceProviderID: entry.ServiceProviderID,
		RedirectURI:           r.Endpoints.SAMLLoginFinishURL(entry.ServiceProviderID).String(),
		Prompt:                prompt,
	}

	return info, nil
}

func (s *UIService) BuildAuthenticationURL(session *SAMLSession) (*url.URL, error) {
	endpoint := s.Endpoints.OAuthEntrypointURL()

	q := endpoint.Query()
	q.Set(queryNameSAMLSessionID, session.ID)
	endpoint.RawQuery = q.Encode()
	return endpoint, nil
}
