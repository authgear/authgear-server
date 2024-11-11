package webapp

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureAuthflowFinishFlowRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern(webapp.AuthflowRouteFinishFlow)
}

type AuthflowFinishFlowHandler struct {
	Controller    *AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
}

func (h *AuthflowFinishFlowHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers AuthflowControllerHandlers
	handlers.Get(func(ctx context.Context, s *webapp.Session, _ *webapp.AuthflowScreenWithFlowResponse) error {
		result, err := h.Controller.Finish(ctx, r, s)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
	h.Controller.HandleWithoutFlow(r.Context(), w, r, &handlers)
}
