package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

func ConfigureNoProjectSSOCallbackRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET").
		WithPathPattern("/noproject/sso/callback")
}

var TemplateWebAuthflowSSOCallbackHTML = template.RegisterHTML(
	"web/authflowv2/sso_callback.html",
	handlerwebapp.Components...,
)

type SSOCallbackHandler struct {
}

func (h *SSOCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
