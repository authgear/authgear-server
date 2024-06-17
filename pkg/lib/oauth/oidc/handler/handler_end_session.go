package handler

import (
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/urlutil"
)

type WebAppURLsProvider interface {
	LogoutURL(redirectURI *url.URL) *url.URL
	SettingsURL() *url.URL
}

type LogoutSessionManager interface {
	Logout(session.ListableSession, http.ResponseWriter) error
}

type CookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
}

type EndSessionHandler struct {
	Config           *config.OAuthConfig
	Endpoints        oidc.EndpointsProvider
	URLs             WebAppURLsProvider
	SessionManager   LogoutSessionManager
	SessionCookieDef session.CookieDef
	Cookies          CookieManager
}

func (h *EndSessionHandler) Handle(s session.ListableSession, req protocol.EndSessionRequest, r *http.Request, rw http.ResponseWriter) error {
	sameSiteStrict, err := h.Cookies.GetCookie(r, h.SessionCookieDef.SameSiteStrictDef)
	if s != nil && err == nil && sameSiteStrict.Value == "true" {
		// Logout directly.
		err := h.SessionManager.Logout(s, rw)
		if err != nil {
			return err
		}
		// Set s to nil and fall through.
		s = nil
	}

	if s != nil {
		endSessionURL := urlutil.WithQueryParamsAdded(
			h.Endpoints.EndSessionEndpointURL(),
			req,
		)
		logoutURL := h.URLs.LogoutURL(endSessionURL)

		http.Redirect(rw, r, logoutURL.String(), http.StatusFound)
		return nil
	}

	redirectURI := req.PostLogoutRedirectURI()
	valid, client := h.validateRedirectURI(redirectURI)
	if !valid {
		// Invalid/empty redirect URI, redirect to home page/settings
		if client != nil && client.ClientURI != "" {
			redirectURI = client.ClientURI
		} else {
			redirectURI = h.URLs.SettingsURL().String()
		}
		http.Redirect(rw, r, redirectURI, http.StatusFound)
		return nil
	}

	if state := req.State(); state != "" {
		uri, err := url.Parse(redirectURI)
		if err != nil {
			return err
		}
		redirectURI = urlutil.WithQueryParamsAdded(uri, map[string]string{"state": state}).String()
	}

	oauth.HTMLRedirect(rw, r, redirectURI)
	return nil
}

func (h *EndSessionHandler) validateRedirectURI(redirectURI string) (valid bool, client *config.OAuthClientConfig) {
	for _, client := range h.Config.Clients {
		for _, uri := range client.PostLogoutRedirectURIs {
			if uri == redirectURI {
				return true, &client
			}
		}
	}
	return false, nil
}
