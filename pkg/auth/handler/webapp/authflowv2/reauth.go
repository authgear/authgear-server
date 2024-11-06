package authflowv2

import (
	"context"
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureAuthflowV2ReauthRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern(AuthflowV2RouteReauth)
}

type AuthflowV2ReauthHandler struct {
	Controller *handlerwebapp.AuthflowController

	AuthflowNavigator handlerwebapp.AuthflowNavigator
}

func (h *AuthflowV2ReauthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	opts := webapp.SessionOptions{
		RedirectURI: h.Controller.RedirectURI(r),
	}

	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		// HandleStartOfFlow used to redirect to the next screen for us.
		// But that redirect was removed.
		// So we need to redirect here.
		// See https://github.com/authgear/authgear-server/issues/3470
		result := &webapp.Result{}
		screen.Navigate(ctx, h.AuthflowNavigator, r, s.ID, result)
		result.WriteResponse(w, r)
		return nil
	})

	h.Controller.HandleStartOfFlow(r.Context(), w, r, opts, authflow.FlowTypeReauth, &handlers, nil)
}
