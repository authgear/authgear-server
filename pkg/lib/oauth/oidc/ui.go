package oidc

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

const queryNameOAuthSessionID = "x_ref"

// UIInfo is a collection of information that is essential to the UI.
type UIInfo struct {
	// ClientID is client_id
	ClientID string
	// RedirectURI is the redirect_uri the UI should redirect to.
	// The redirect_uri in the URL has lower precedence.
	// The rationale for this is if the end-user bookmarked the
	// authorization URL in the browser, redirect to the app is
	// possible.
	RedirectURI string
	// Prompt is the resolved prompt with prompt, max_age, and id_token_hint taken into account.
	Prompt []string
	// UILocales is ui_locales.
	UILocales string
	// UserIDHint is for reauthentication.
	UserIDHint string
	// CanUseIntentReauthenticate is for reauthentication.
	CanUseIntentReauthenticate bool
	// State is the state parameter
	State string
	// XState is the x_state parameter
	XState string
	// Page is the x_page parameter
	Page string
	// SuppressIDPSessionCookie is the x_suppress_idp_session_cookie and x_sso_enabled parameter.
	SuppressIDPSessionCookie bool
	// OAuthProviderAlias is the x_oauth_provider_alias parameter.
	OAuthProviderAlias string
	// LoginHint is the OIDC login_hint parameter.
	LoginHint string
	// IDTokenHint is the OIDC id_token_hint parameter.
	IDTokenHint string
}

func (i *UIInfo) ToUIParam() uiparam.T {
	return uiparam.T{
		ClientID:  i.ClientID,
		Prompt:    i.Prompt,
		State:     i.State,
		XState:    i.XState,
		UILocales: i.UILocales,
	}
}

type UIInfoByProduct struct {
	IDToken        jwt.Token
	SIDSession     session.ListableSession
	IDTokenHintSID string
}

type UIInfoResolverPromptResolver interface {
	ResolvePrompt(r protocol.AuthorizationRequest, sidSession session.ListableSession) (prompt []string)
}

type UIInfoResolverIDTokenHintResolver interface {
	ResolveIDTokenHint(client *config.OAuthClientConfig, r protocol.AuthorizationRequest) (idToken jwt.Token, sidSession session.ListableSession, err error)
}

type UIInfoResolverCookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
	ClearCookie(def *httputil.CookieDef) *http.Cookie
}

type UIInfoClientResolver interface {
	ResolveClient(clientID string) *config.OAuthClientConfig
}

type UIInfoResolver struct {
	Config              *config.OAuthConfig
	EndpointsProvider   oauth.EndpointsProvider
	PromptResolver      UIInfoResolverPromptResolver
	IDTokenHintResolver UIInfoResolverIDTokenHintResolver
	Clock               clock.Clock
	Cookies             UIInfoResolverCookieManager
	ClientResolver      UIInfoClientResolver
}

func (r *UIInfoResolver) SetAuthenticationInfoInQuery(redirectURI string, e *authenticationinfo.Entry) string {
	consentURL := r.EndpointsProvider.ConsentEndpointURL()

	u, err := url.Parse(redirectURI)
	if err != nil {
		panic(err)
	}

	// When redirectURI is consentURL, it will have client_id, redirect_uri, state in it,
	// so we have to compare them WITHOUT query nor fragment.
	// When we are not redirecting to consentURL, we do not set code.
	equalWithoutQueryNorFragment := u.Scheme == consentURL.Scheme && u.Host == consentURL.Host && u.Path == consentURL.Path
	if !equalWithoutQueryNorFragment {
		return redirectURI
	}

	q := u.Query()
	q.Set("code", e.ID)
	u.RawQuery = q.Encode()
	return u.String()
}

func (r *UIInfoResolver) GetAuthenticationInfoID(req *http.Request) (string, bool) {
	code := req.FormValue("code")
	if code != "" {
		return code, true
	}
	return "", false
}

func (r *UIInfoResolver) GetOAuthSessionID(req *http.Request, urlQuery string) (string, bool) {
	if q, err := url.ParseQuery(urlQuery); err == nil {
		id := q.Get(queryNameOAuthSessionID)
		if id != "" {
			return id, true
		}
	}

	id := req.URL.Query().Get(queryNameOAuthSessionID)
	if id != "" {
		return id, true
	}
	return "", false
}

func (r *UIInfoResolver) GetOAuthSessionIDLegacy(req *http.Request, urlQuery string) (string, bool) {
	if q, err := url.ParseQuery(urlQuery); err == nil {
		id := q.Get(queryNameOAuthSessionID)
		if id != "" {
			return id, true
		}
	}

	id := req.URL.Query().Get(queryNameOAuthSessionID)
	if id != "" {
		return id, true
	}
	cookie, err := r.Cookies.GetCookie(req, oauthsession.UICookieDef)
	if err == nil {
		return cookie.Value, true
	}
	return "", false
}

func (r *UIInfoResolver) RemoveOAuthSessionID(w http.ResponseWriter, req *http.Request) {
	// Remove from http.Request.URL
	urlQuery := req.URL.Query()
	urlQuery.Del(queryNameOAuthSessionID)
	reqURL := *req.URL
	reqURL.RawQuery = urlQuery.Encode()
	req.URL = &reqURL

	// Remove from cookies
	httputil.UpdateCookie(w, r.Cookies.ClearCookie(oauthsession.UICookieDef))
}

func (r *UIInfoResolver) ResolveForUI(req protocol.AuthorizationRequest) (*UIInfo, error) {
	client := r.ClientResolver.ResolveClient(req.ClientID())
	if client == nil {
		return nil, fmt.Errorf("client not found: %v", req.ClientID())
	}

	uiInfo, _, err := r.ResolveForAuthorizationEndpoint(client, req)
	return uiInfo, err
}

func (r *UIInfoResolver) ResolveForAuthorizationEndpoint(
	client *config.OAuthClientConfig,
	req protocol.AuthorizationRequest,
) (*UIInfo, *UIInfoByProduct, error) {
	redirectURI := r.EndpointsProvider.ConsentEndpointURL()

	// Add client_id, redirect_uri, state to URL as hint when oauth session expires / not found
	q := redirectURI.Query()
	q.Add("client_id", req.ClientID())
	q.Add("redirect_uri", req.RedirectURI())
	if state := req.State(); state != "" {
		q.Add("state", state)
	}
	redirectURI.RawQuery = q.Encode()

	idToken, sidSession, err := r.IDTokenHintResolver.ResolveIDTokenHint(client, req)
	if err != nil {
		return nil, nil, err
	}

	prompt := r.PromptResolver.ResolvePrompt(req, sidSession)

	var idTokenHintSID string
	if sidSession != nil {
		idTokenHintSID = EncodeSID(sidSession)
	}

	var userIDHint string
	var canUseIntentReauthenticate bool
	if idToken != nil {
		userIDHint = idToken.Subject()
		if tv := idToken.Expiration(); !tv.IsZero() && tv.Unix() != 0 {
			now := r.Clock.NowUTC().Truncate(time.Second)
			tv = tv.Truncate(time.Second)
			if now.Before(tv) && sidSession != nil {
				switch sidSession.SessionType() {
				case session.TypeIdentityProvider:
					canUseIntentReauthenticate = true
				case session.TypeOfflineGrant:
					if offlineGrant, ok := sidSession.(*oauth.OfflineGrant); ok {
						if offlineGrant.HasAllScopes(req.ClientID(), []string{oauth.FullAccessScope}) {
							canUseIntentReauthenticate = true
						}
					}
				}
			}
		}
	}

	loginIDHint, _ := req.LoginHint()

	idTokenHint, _ := req.IDTokenHint()

	info := &UIInfo{
		ClientID:                   req.ClientID(),
		RedirectURI:                redirectURI.String(),
		Prompt:                     prompt,
		UserIDHint:                 userIDHint,
		CanUseIntentReauthenticate: canUseIntentReauthenticate,
		State:                      req.State(),
		XState:                     req.XState(),
		Page:                       req.Page(),
		SuppressIDPSessionCookie:   req.SuppressIDPSessionCookie(),
		OAuthProviderAlias:         req.OAuthProviderAlias(),
		UILocales:                  req.UILocalesRaw(),
		LoginHint:                  loginIDHint,
		IDTokenHint:                idTokenHint,
	}
	byProduct := &UIInfoByProduct{
		IDToken:        idToken,
		SIDSession:     sidSession,
		IDTokenHintSID: idTokenHintSID,
	}
	return info, byProduct, nil
}

type UIURLBuilderAuthUIEndpointsProvider interface {
	OAuthEntrypointURL() *url.URL
	SettingsChangePasswordURL() *url.URL
}

type UIURLBuilder struct {
	Endpoints UIURLBuilderAuthUIEndpointsProvider
}

func (b *UIURLBuilder) BuildAuthenticationURL(client *config.OAuthClientConfig, r protocol.AuthorizationRequest, e *oauthsession.Entry) (*url.URL, error) {
	var endpoint *url.URL
	if client != nil && client.CustomUIURI != "" {
		var err error
		endpoint, err = BuildCustomUIEndpoint(client.CustomUIURI)
		if err != nil {
			return nil, ErrInvalidCustomURI.Errorf("invalid custom ui uri: %w", err)
		}
	} else {
		endpoint = b.Endpoints.OAuthEntrypointURL()
	}

	q := endpoint.Query()
	q.Set(queryNameOAuthSessionID, e.ID)
	q.Set("client_id", r.ClientID())
	q.Set("redirect_uri", r.RedirectURI())
	if r.ColorScheme() != "" {
		q.Set("x_color_scheme", r.ColorScheme())
	}
	if len(r.UILocales()) > 0 {
		q.Set("ui_locales", strings.Join(r.UILocales(), " "))
	}
	if r.State() != "" {
		q.Set("state", r.State())
	}
	if r.XState() != "" {
		q.Set("x_state", r.XState())
	}
	endpoint.RawQuery = q.Encode()

	return endpoint, nil
}

func BuildCustomUIEndpoint(base string) (*url.URL, error) {
	customUIURL, err := url.Parse(base)
	if err != nil {
		return nil, err
	}

	return customUIURL, nil
}

func (b *UIURLBuilder) BuildSettingsActionURL(client *config.OAuthClientConfig, r protocol.AuthorizationRequest, e *oauthsession.Entry, redirectURI *url.URL) (*url.URL, error) {
	switch r.SettingsAction() {
	case "change_password":
		endpoint := b.Endpoints.SettingsChangePasswordURL()
		q := endpoint.Query()
		q.Set(queryNameOAuthSessionID, e.ID)
		q.Set("client_id", r.ClientID())
		q.Set("redirect_uri", r.RedirectURI())
		if r.ColorScheme() != "" {
			q.Set("x_color_scheme", r.ColorScheme())
		}
		if len(r.UILocales()) > 0 {
			q.Set("ui_locales", strings.Join(r.UILocales(), " "))
		}
		if r.State() != "" {
			q.Set("state", r.State())
		}
		if r.XState() != "" {
			q.Set("x_state", r.XState())
		}
		endpoint.RawQuery = q.Encode()
		return endpoint, nil
	default:
		return nil, ErrInvalidSettingsAction.New("invalid settings action")
	}
}
