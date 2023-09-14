package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebTesterHTML = template.RegisterHTML(
	"web/tester.html",
	components...,
)

func ConfigureTesterRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET", "POST").
		WithPathPattern("/tester")
}

type TesterHandler struct {
	ControllerFactory ControllerFactory
	EndpointsProvider oauth.EndpointsProvider
}

func (h *TesterHandler) triggerAuth(token string, w http.ResponseWriter, r *http.Request) error {
	authEndpoint := h.EndpointsProvider.AuthorizeEndpointURL()
	// TODO
}

func (h *TesterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	ctrl.Get(func() error {
		token := r.URL.Query().Get("token")
		if token != "" {
			return h.triggerAuth(token, w, r)
		}
		return nil
	})

}

var _ http.Handler = &TesterHandler{}
