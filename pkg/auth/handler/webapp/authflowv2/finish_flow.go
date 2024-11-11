package authflowv2

import (
	"context"
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureAuthflowv2FinishFlowRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern(AuthflowV2RouteFinishFlow)
}

type AuthflowV2FinishFlowHandler struct {
	Controller    *handlerwebapp.AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      handlerwebapp.Renderer
}

func (h *AuthflowV2FinishFlowHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
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
