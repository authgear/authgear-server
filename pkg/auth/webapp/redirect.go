package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func GetRedirectURI(r *http.Request, trustProxy bool, defaultURI string) string {
	redirectURI, err := httputil.GetRedirectURI(r, trustProxy)
	if err != nil {
		return defaultURI
	}
	return redirectURI
}

func DefaultPostLoginRedirectURI(uiConfig *config.UIConfig) string {
	if uiConfig != nil && uiConfig.DefaultRedirectURI != "" {
		return uiConfig.DefaultRedirectURI
	}
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
