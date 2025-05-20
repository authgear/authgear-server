package webapp

import (
	"context"
	"net/http"

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
	Components...,
)

type NoProjectSSOCallbackHandlerOAuthStateStore interface {
	RecoverState(ctx context.Context, stateToken string) (state *webappoauth.WebappOAuthState, err error)
}

type NoProjectSSOCallbackHandler struct {
	OAuthStateStore NoProjectSSOCallbackHandlerOAuthStateStore
}

func (h *NoProjectSSOCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
