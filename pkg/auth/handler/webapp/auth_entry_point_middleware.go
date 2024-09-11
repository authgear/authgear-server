package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateDirectAccessDisable = template.RegisterHTML(
	"web/authflowv2/direct_access_disable.html",
	DirectAccessDisableComponents...,
)

type AuthEntryPointMiddleware struct {
	BaseViewModel       *viewmodels.BaseViewModeler
	Renderer            Renderer
	AppHostSuffixes     config.AppHostSuffixes
	TrustProxy          config.TrustProxy
	OAuthConfig         *config.OAuthConfig
	UIConfig            *config.UIConfig
	OAuthClientResolver WebappOAuthClientResolver
}

func (m *AuthEntryPointMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := session.GetUserID(r.Context())
		webSession := webapp.GetSession(r.Context())

		fromAuthzEndpoint := false
		if webSession != nil {
			// stay in the auth entry point if login is triggered by authz endpoint
			fromAuthzEndpoint = webSession.OAuthSessionID != "" || webSession.SAMLSessionID != ""
		}

		host := httputil.GetHost(r, bool(m.TrustProxy))
		isDefaultDomain := m.AppHostSuffixes.CheckIsDefaultDomain(host)

		if userID != nil && !fromAuthzEndpoint && !m.UIConfig.DirectAccessDisabled {
			defaultRedirectURI := webapp.DerivePostLoginRedirectURIFromRequest(r, m.OAuthClientResolver, m.UIConfig)
			redirectURI := webapp.GetRedirectURI(r, bool(m.TrustProxy), defaultRedirectURI)

			http.Redirect(w, r, redirectURI, http.StatusFound)
		} else if (userID == nil && !fromAuthzEndpoint && isDefaultDomain) || m.UIConfig.DirectAccessDisabled {
			m.renderBlocked(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func (m *AuthEntryPointMiddleware) renderBlocked(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	baseViewModel := m.BaseViewModel.ViewModel(r, w)
	viewmodels.Embed(data, baseViewModel)
	m.Renderer.RenderHTML(w, r, TemplateDirectAccessDisable, data)
}
