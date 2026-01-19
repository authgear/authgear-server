package webapp

import (
	"log/slog"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var AuthEntryPointMiddlewareLogger = slogutil.NewLogger("auth_entry_point_middleware")

var TemplateDirectAccessDisable = template.RegisterHTML(
	"web/authflowv2/direct_access_disabled.html",
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
		ctx := r.Context()
		userID := session.GetUserID(ctx)
		webSession := webapp.GetSession(ctx)

		fromAuthzEndpoint := false
		if webSession != nil {
			// stay in the auth entry point if login is triggered by authz endpoint
			fromAuthzEndpoint = webSession.OAuthSessionID != "" || webSession.SAMLSessionID != ""
		}

		host := httputil.GetHost(r, bool(m.TrustProxy))
		isDefaultDomain := m.AppHostSuffixes.CheckIsDefaultDomain(host)
		directAccessDisabled := isDefaultDomain || m.UIConfig.DirectAccessDisabled

		if userID != nil && !fromAuthzEndpoint {
			defaultRedirectURI := webapp.DerivePostLoginRedirectURIFromRequest(r, m.OAuthClientResolver, m.UIConfig)
			redirectURI := webapp.GetRedirectURI(r, bool(m.TrustProxy), defaultRedirectURI)

			http.Redirect(w, r, redirectURI, http.StatusFound)
		} else if userID == nil && !fromAuthzEndpoint && directAccessDisabled {
			logger := AuthEntryPointMiddlewareLogger.GetLogger(ctx)
			var webSessionID string
			var oauthSessionID string
			var samlSessionID string
			var cookies []string = []string{}
			if webSession != nil {
				webSessionID = webSession.ID
				oauthSessionID = webSession.OAuthSessionID
				samlSessionID = webSession.SAMLSessionID
			}
			allCookies := r.Cookies()
			for _, cookie := range allCookies {
				if cookie.Value != "" {
					// Log all existing cookies
					cookies = append(cookies, cookie.Name)
				}
			}
			logger.WithSkipStackTrace().WithSkipLogging().Warn(ctx, "auth direct access blocked",
				slog.String("web_session_id", webSessionID),
				slog.String("oauth_session_id", oauthSessionID),
				slog.String("saml_session_id", samlSessionID),
				slog.Any("cookies", cookies),
			)
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
