package samlsession

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlerror"
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

	// login_hint resolved from <Subject>
	LoginHint string
}

func (i *SAMLUIInfo) ToUIParam() uiparam.T {
	return uiparam.T{
		Prompt: i.Prompt,
	}
}

type UIServiceAuthUIEndpointsProvider interface {
	OAuthEntrypointURL() *url.URL
	SAMLLoginFinishURL() *url.URL
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

func (r *UIService) ResolveUIInfo(sp *config.SAMLServiceProviderConfig, entry *SAMLSessionEntry) (
	info *SAMLUIInfo, showUI bool, err error) {
	var prompt []string
	authnRequest, authnRequestExist := entry.AuthnRequest()
	switch {
	case !authnRequestExist:
		// This is an Idp-Initiated flow, allow user to select_account or login
		prompt = []string{}
		showUI = true
	case authnRequest.GetIsPassive() == false && authnRequest.GetForceAuthn() == false:
		prompt = []string{}
		showUI = true
	case authnRequest.GetIsPassive() == false && authnRequest.GetForceAuthn() == true:
		prompt = []string{"login"}
		showUI = true
	case authnRequest.GetIsPassive() == true && authnRequest.GetForceAuthn() == false:
		// prompt=none
		showUI = false
	default:
		// Other cases should be blocked in request validation stage.
		// It is an unexpected error if it reaches here
		return nil, false, fmt.Errorf("unexpected: IsPassive=%v and ForceAuthn=%v",
			authnRequest.GetIsPassive(),
			authnRequest.GetForceAuthn())
	}

	var loginHintStr string
	if authnRequestExist && authnRequest.Subject != nil && authnRequest.Subject.NameID != nil {
		nameID := authnRequest.Subject.NameID
		loginHint := &oauth.LoginHint{
			Type:    oauth.LoginHintTypeLoginID,
			Enforce: true,
		}
		switch sp.NameIDFormat {
		case config.SAMLNameIDFormatEmailAddress:
			loginHint.LoginIDEmail = nameID.Value
		case config.SAMLNameIDFormatUnspecified:
			switch sp.NameIDAttributePointer {
			case "/email":
				loginHint.LoginIDEmail = nameID.Value
			case "/phone_number":
				loginHint.LoginIDPhone = nameID.Value
			case "/preferred_username":
				loginHint.LoginIDUsername = nameID.Value
			default:
				return nil, false, &samlerror.InvalidRequestError{
					Field:  "Subject",
					Reason: "Using <Subject> in <AuthnRequest> is only supported when nameid_attribute_pointer is '/email', '/phone_number' or 'preferred_username'",
				}
			}
		default:
			panic(fmt.Errorf("unknown nameid format %v", sp.NameIDFormat))
		}
		loginHintStr = loginHint.String()
	}

	info = &SAMLUIInfo{
		SAMLServiceProviderID: entry.ServiceProviderID,
		RedirectURI:           r.Endpoints.SAMLLoginFinishURL().String(),
		Prompt:                prompt,
		LoginHint:             loginHintStr,
	}

	return info, showUI, nil
}

func (s *UIService) BuildAuthenticationURL(session *SAMLSession) (*url.URL, error) {
	endpoint := s.Endpoints.OAuthEntrypointURL()

	q := endpoint.Query()
	q.Set(queryNameSAMLSessionID, session.ID)
	endpoint.RawQuery = q.Encode()
	return endpoint, nil
}
