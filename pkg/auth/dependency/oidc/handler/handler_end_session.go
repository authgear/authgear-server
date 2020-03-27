package handler

import (
	"net/http"
	"net/url"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oidc"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oidc/protocol"
	"github.com/skygeario/skygear-server/pkg/core/config"
	coreurl "github.com/skygeario/skygear-server/pkg/core/url"
)

// TODO(oidc): write tests

type EndSessionHandler struct {
	Clients            []config.OAuthClientConfiguration
	EndSessionEndpoint oidc.EndSessionEndpointProvider
	LogoutEndpoint     oauth.LogoutEndpointProvider
	SettingsEndpoint   oauth.SettingsEndpointProvider
}

func (h *EndSessionHandler) Handle(s auth.AuthSession, req protocol.EndSessionRequest, r *http.Request, rw http.ResponseWriter) error {
	if s != nil {
		endSessionURI := coreurl.WithQueryParamsAdded(
			h.EndSessionEndpoint.EndSessionEndpointURI(),
			req,
		)
		logoutURI := coreurl.WithQueryParamsAdded(
			h.LogoutEndpoint.LogoutEndpointURI(),
			map[string]string{"redirect_uri": endSessionURI.String()},
		)

		http.Redirect(rw, r, logoutURI.String(), http.StatusFound)
		return nil
	}

	redirectURI := req.PostLogoutRedirectURI()
	valid, client := h.validateRedirectURI(redirectURI)
	if !valid {
		// Invalid/empty redirect URI, redirect to home page/settings
		if client != nil && client.ClientURI() != "" {
			redirectURI = client.ClientURI()
		} else {
			redirectURI = h.SettingsEndpoint.SettingsEndpointURI().String()
		}
		http.Redirect(rw, r, redirectURI, http.StatusFound)
		return nil
	}

	if state := req.State(); state != "" {
		uri, err := url.Parse(redirectURI)
		if err != nil {
			return err
		}
		redirectURI = coreurl.WithQueryParamsAdded(uri, map[string]string{"state": state}).String()
	}

	http.Redirect(rw, r, redirectURI, http.StatusFound)
	return nil
}

func (h *EndSessionHandler) validateRedirectURI(redirectURI string) (valid bool, client config.OAuthClientConfiguration) {
	for _, client := range h.Clients {
		for _, uri := range client.PostLogoutRedirectURIs() {
			if uri == redirectURI {
				return true, client
			}
		}
	}
	return false, nil
}
