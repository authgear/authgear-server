package webapp

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureCSRFErrorInstructionRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/errors/cookie_instruction")
}

type CSRFErrorInstructionHandler struct {
	Controller    *AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
}

func (h *CSRFErrorInstructionHandler) GetData(w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	userAgent := r.UserAgent()
	device, _ := model.GetRecognizedMobileDevice(userAgent)
	data["Device"] = device

	return data, nil
}

func (h *CSRFErrorInstructionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers AuthflowControllerHandlers
	handlers.Get(func(ctx context.Context, _ *webapp.Session, _ *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateCSRFErrorInstructionHTML, data)
		return nil
	})
	h.Controller.HandleWithoutSession(r.Context(), w, r, &handlers)
}
