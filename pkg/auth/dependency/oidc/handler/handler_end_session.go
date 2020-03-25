package handler

import (
	"net/http"
	"net/url"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oidc/protocol"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

// TODO(oidc): write tests

type endSessionManager interface {
	Logout(session auth.AuthSession, rw http.ResponseWriter) error
}

type EndSessionHandler struct {
	Clients          []config.OAuthClientConfiguration
	Sessions         endSessionManager
	SettingsEndpoint oauth.SettingsEndpointProvider
}

func (h *EndSessionHandler) Handle(s auth.AuthSession, req protocol.EndSessionRequest, r *http.Request, rw http.ResponseWriter) error {
	redirectURI := req.PostLogoutRedirectURI()
	if !h.validateRedirectURI(s, redirectURI) {
		// Invalid/empty redirect URI, redirect to settings to prompt user
		http.Redirect(rw, r, h.SettingsEndpoint.SettingsEndpointURI().String(), http.StatusFound)
		return nil
	}

	if state := req.State(); state != "" {
		uri, err := url.Parse(redirectURI)
		if err != nil {
			return err
		}
		query := uri.Query()
		query.Set("state", state)
		uri.RawQuery = query.Encode()
		redirectURI = uri.String()
	}

	if s != nil {
		// TODO(oidc): handle id_token_hint

		err := h.Sessions.Logout(s, rw)
		if err != nil {
			return err
		}
	}

	http.Redirect(rw, r, redirectURI, http.StatusFound)
	return nil
}

func (h *EndSessionHandler) validateRedirectURI(s auth.AuthSession, redirectURI string) bool {
	for _, client := range h.Clients {
		if s != nil && s.GetClientID() != "" && client.ClientID() != s.GetClientID() {
			continue
		}
		for _, uri := range client.PostLogoutRedirectURIs() {
			if uri == redirectURI {
				return true
			}
		}
	}
	return false
}
