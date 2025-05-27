package webapp

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/endpoints"
	"github.com/authgear/authgear-server/pkg/lib/webappoauth"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

func ConfigureNoProjectSSOCallbackRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/noproject/sso/oauth2/callback")
}

var TemplateWebAuthflowSSOCallbackHTML = template.RegisterHTML(
	"web/authflowv2/sso_callback.html",
	Components...,
)

type NoProjectSSOCallbackHandlerOAuthStateStore interface {
	RecoverState(ctx context.Context, stateToken string) (state *webappoauth.WebappOAuthState, err error)
}

type NoProjectSSOCallbackHandler struct {
	ConfigSource    *configsource.ConfigSource
	OAuthStateStore NoProjectSSOCallbackHandlerOAuthStateStore
}

func (h *NoProjectSSOCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	stateToken := r.FormValue("state")
	state, err := h.OAuthStateStore.RecoverState(ctx, stateToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	publicOrigin := ""
	err = h.ConfigSource.ContextResolver.ResolveContext(ctx, state.AppID, func(ctx context.Context, appCtx *config.AppContext) error {
		publicOrigin = appCtx.Config.AppConfig.HTTP.PublicOrigin
		return nil
	})
	if err != nil {
		panic(fmt.Errorf("failed to resolve public origin of app id: %s %w", state.AppID, err))
	}

	publicOriginURL, err := url.Parse(publicOrigin)
	if err != nil {
		panic(fmt.Errorf("unexpected: public origin of app %s is not a valid url. %w", state.AppID, err))
	}
	oauthEndpoints := endpoints.NewOAuthEndpoints(publicOriginURL)
	redirectURL := oauthEndpoints.SSOCallbackURL(state.ProviderAlias)
	redirectURL.RawQuery = r.URL.RawQuery

	// Use 307 so that method and body is kept
	http.Redirect(w, r, redirectURL.String(), http.StatusTemporaryRedirect)
}
