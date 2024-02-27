package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/iawaknahc/originmatcher"
)

func GetRedirectURI(r *http.Request, trustProxy bool, defaultURI string) string {
	redirectURI, err := httputil.GetRedirectURI(r, trustProxy)
	if err != nil {
		return defaultURI
	}
	return redirectURI
}

type OAuthClientResolver interface {
	ResolveClient(clientID string) *config.OAuthClientConfig
}

func DeriveSettingsRedirectURIFromRequest(r *http.Request, clientResolver OAuthClientResolver, defaultURI string) string {
	// 1. Redirect URL in query param (must be whitelisted)
	// 2. Default redirect URL
	// 3. `/settings`
	redirectURIFromQuery := func() string {
		clientID := r.URL.Query().Get("client_id")
		redirectURI := r.URL.Query().Get("redirect_uri")
		if clientID == "" {
			return ""
		}
		client := clientResolver.ResolveClient(clientID)
		if client == nil {
			return ""
		}

		allowed := true
		matcher, err := originmatcher.New(client.SettingsRedirectURIOrigins)
		if err != nil {
			return ""
		}

		if matcher.MatchOrigin(redirectURI) {
			allowed = true
		}

		// 1. Redirect URL in query param (must be whitelisted)
		if allowed && redirectURI != "" {
			return redirectURI
		}

		return ""
	}()

	if redirectURIFromQuery != "" {
		return redirectURIFromQuery
	}

	// 2. Default redirect URL
	if defaultURI != "" {
		return defaultURI
	}

	// 3. `/settings`
	return "/settings"
}

func DerivePostLoginRedirectURIFromRequest(r *http.Request, clientResolver OAuthClientResolver, uiConfig *config.UIConfig) string {
	// 1. Redirect URL in query param (must be whitelisted)
	// 2. Default redirect URL of the client
	// 3. Post-login URL
	// 4. `/settings`
	redirectURIFromQuery := func() string {
		clientID := r.URL.Query().Get("client_id")
		redirectURI := r.URL.Query().Get("redirect_uri")
		if clientID == "" {
			return ""
		}
		client := clientResolver.ResolveClient(clientID)
		if client == nil {
			return ""
		}

		allowed := false
		for _, u := range client.RedirectURIs {
			if u == redirectURI {
				allowed = true
				break
			}
		}

		// 1. Redirect URL in query param (must be whitelisted)
		if allowed && redirectURI != "" {
			return redirectURI
		}

		// 2. Default redirect URL of the client
		return client.DefaultRedirectURI()
	}()

	if redirectURIFromQuery != "" {
		return redirectURIFromQuery
	}

	// 3. Post-login URL
	if uiConfig != nil && uiConfig.DefaultRedirectURI != "" {
		return uiConfig.DefaultRedirectURI
	}

	// 4. `/settings`
	return "/settings"
}

func ResolvePostLogoutRedirectURI(client *config.OAuthClientConfig, givenPostLogoutRedirectURI string, uiConfig *config.UIConfig) string {
	if client != nil && givenPostLogoutRedirectURI != "" {
		for _, v := range client.PostLogoutRedirectURIs {
			if v == givenPostLogoutRedirectURI {
				return givenPostLogoutRedirectURI
			}
		}
	}

	if uiConfig != nil && uiConfig.DefaultPostLogoutRedirectURI != "" {
		return uiConfig.DefaultPostLogoutRedirectURI
	}

	return "/login"
}

func ResolveClientURI(client *config.OAuthClientConfig, uiConfig *config.UIConfig) string {
	if client != nil && client.ClientURI != "" {
		return client.ClientURI
	}
	if uiConfig != nil && uiConfig.DefaultClientURI != "" {
		return uiConfig.DefaultClientURI
	}
	return ""
}
