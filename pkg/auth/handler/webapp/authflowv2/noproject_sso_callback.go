package authflowv2

import (
	"context"
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/lib/webappoauth"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

func ConfigureNoProjectSSOCallbackRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET").
		WithPathPattern("/noproject/sso/oauth2/callback")
}

var TemplateWebAuthflowSSOCallbackHTML = template.RegisterHTML(
	"web/authflowv2/sso_callback.html",
	handlerwebapp.Components...,
)

type NoProjectSSOCallbackHandlerOAuthStateStore interface {
	RecoverState(ctx context.Context, stateToken string) (state *webappoauth.WebappOAuthState, err error)
}

type SSOCallbackHandler struct {
	OAuthStateStore NoProjectSSOCallbackHandlerOAuthStateStore
}

func (h *SSOCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	stateToken := r.FormValue("state")
	_, err := h.OAuthStateStore.RecoverState(r.Context(), stateToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// TODO
}
