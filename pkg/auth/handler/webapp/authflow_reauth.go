package webapp

import (
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureAuthflowReauthRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern(webapp.AuthflowRouteReauth)
}

type AuthflowReauthHandler struct {
	Controller *AuthflowController
}

func (h *AuthflowReauthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flowName := "default"
	opts := webapp.SessionOptions{
		RedirectURI: h.Controller.RedirectURI(r),
	}

	var handlers AuthflowControllerHandlers
	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		return fmt.Errorf("unreachable")
	})

	h.Controller.HandleStartOfFlow(w, r, opts, authflow.FlowReference{
		Type: authflow.FlowTypeReauth,
		Name: flowName,
	}, &handlers)
}
