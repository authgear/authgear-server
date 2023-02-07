package oidc

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/jwt"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

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
	// UserIDHint is for reauthentication.
	UserIDHint string
	// CanUseIntentReauthenticate is for reauthentication.
	CanUseIntentReauthenticate bool
	// State is the state parameter
	State string
	// Page is the x_page parameter
	Page string
	// SuppressIDPSessionCookie is the x_suppress_idp_session_cookie and x_sso_enabled parameter.
	SuppressIDPSessionCookie bool
	// OAuthProviderAlias is the x_oauth_provider_alias parameter.
	OAuthProviderAlias string
	// LoginHint is the OIDC login_hint parameter.
	LoginHint string
}

type UIInfoByProduct struct {
	IDToken        jwt.Token
	SIDSession     session.Session
	IDTokenHintSID string
}

type UIInfoResolverPromptResolver interface {
	ResolvePrompt(r protocol.AuthorizationRequest, sidSession session.Session) (prompt []string)
}

type UIInfoResolverIDTokenHintResolver interface {
	ResolveIDTokenHint(client *config.OAuthClientConfig, r protocol.AuthorizationRequest) (idToken jwt.Token, sidSession session.Session, err error)
}

type UIInfoResolver struct {
	Config              *config.OAuthConfig
	EndpointsProvider   oauth.EndpointsProvider
	PromptResolver      UIInfoResolverPromptResolver
	IDTokenHintResolver UIInfoResolverIDTokenHintResolver
	Clock               clock.Clock
}

func (r *UIInfoResolver) ResolveForUI(req protocol.AuthorizationRequest) (*UIInfo, error) {
	client, ok := r.Config.GetClient(req.ClientID())
	if !ok {
		return nil, fmt.Errorf("client not found: %v", req.ClientID())
	}

	uiInfo, _, err := r.ResolveForAuthorizationEndpoint(client, req)
	return uiInfo, err
}

func (r *UIInfoResolver) ResolveForAuthorizationEndpoint(
	client *config.OAuthClientConfig,
	req protocol.AuthorizationRequest,
) (*UIInfo, *UIInfoByProduct, error) {
	redirectURI := r.EndpointsProvider.ConsentEndpointURL().String()

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
						if slice.ContainsString(offlineGrant.Scopes, oauth.FullAccessScope) {
							canUseIntentReauthenticate = true
						}
					}
				}
			}
		}
	}

	loginIDHint, _ := req.LoginHint()

	info := &UIInfo{
		ClientID:                   req.ClientID(),
		RedirectURI:                redirectURI,
		Prompt:                     prompt,
		UserIDHint:                 userIDHint,
		CanUseIntentReauthenticate: canUseIntentReauthenticate,
		State:                      req.State(),
		Page:                       req.Page(),
		SuppressIDPSessionCookie:   req.SuppressIDPSessionCookie(),
		OAuthProviderAlias:         req.OAuthProviderAlias(),
		LoginHint:                  loginIDHint,
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
}

type UIURLBuilder struct {
	Endpoints UIURLBuilderAuthUIEndpointsProvider
}

func (b *UIURLBuilder) Build(client *config.OAuthClientConfig, r protocol.AuthorizationRequest) (*url.URL, error) {
	var endpoint *url.URL
	if client != nil && client.CustomUIURI != "" {
		var err error
		endpoint, err = BuildCustomUIEndpoint(client.CustomUIURI, r.CustomUIQuery())
		if err != nil {
			return nil, ErrInvalidCustomURI.Errorf("invalid custom ui uri: %w", err)
		}
	} else {
		endpoint = b.Endpoints.OAuthEntrypointURL()
	}

	q := endpoint.Query()
	q.Set("client_id", r.ClientID())
	q.Set("redirect_uri", r.RedirectURI())
	if r.ColorScheme() != "" {
		q.Set("x_color_scheme", r.ColorScheme())
	}
	if len(r.UILocales()) > 0 {
		q.Set("ui_locales", strings.Join(r.UILocales(), " "))
	}
	endpoint.RawQuery = q.Encode()

	return endpoint, nil
}

func BuildCustomUIEndpoint(base string, query string) (*url.URL, error) {
	customUIURL, err := url.Parse(base)
	if err != nil {
		return nil, err
	}

	q := customUIURL.Query()

	// Assign query from the SDK to the url
	queryFromSDK, err := url.ParseQuery(query)
	if err != nil {
		return nil, err
	}
	for key, values := range queryFromSDK {
		for idx, val := range values {
			if idx == 0 {
				q.Set(key, val)
			} else {
				q.Add(key, val)
			}
		}
	}
	customUIURL.RawQuery = q.Encode()

	return customUIURL, nil
}
