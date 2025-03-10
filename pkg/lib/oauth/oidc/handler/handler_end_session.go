package handler

import (
	"context"
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
	Logout(ctx context.Context, sessionBase session.SessionBase, w http.ResponseWriter) ([]session.ListableSession, error)
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

func (h *EndSessionHandler) Handle(ctx context.Context, s session.ResolvedSession, req protocol.EndSessionRequest, r *http.Request, rw http.ResponseWriter) error {
	sameSiteStrict, err := h.Cookies.GetCookie(r, h.SessionCookieDef.SameSiteStrictDef)
	if s != nil && err == nil && sameSiteStrict.Value == "true" {
		// Logout directly.
		// TODO(SAML): Logout affected saml service providers
		_, err := h.SessionManager.Logout(ctx, s, rw)
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

	redirectURIURL, err := url.Parse(redirectURI)
	if err != nil {
		panic(err)
	}

	writeResponseOptions := oauth.WriteResponseOptions{
		RedirectURI:  redirectURIURL,
		ResponseMode: "query",
		UseHTTP200:   client.UseHTTP200(),
		Response:     make(map[string]string),
	}
	oauth.WriteResponse(rw, r, writeResponseOptions)
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
