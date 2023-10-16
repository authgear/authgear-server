package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebAuthflowAccountStatusHTML = template.RegisterHTML(
	"web/authflow_account_status.html",
	components...,
)

func ConfigureAuthflowAccountStatusRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(webapp.AuthflowRouteAccountStatus)
}

type AuthflowAccountStatusHandler struct {
	Controller    *AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
}

func (h *AuthflowAccountStatusHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModel(r, w)
	if screenData, ok := screen.StateTokenFlowResponse.Action.Data.(declarative.NodeDidCheckAccountStatusData); ok {
		baseViewModel.SetError(screenData.Error)
	}

	viewmodels.Embed(data, baseViewModel)

	return data, nil
}

func (h *AuthflowAccountStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers AuthflowControllerHandlers
	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowAccountStatusHTML, data)
		return nil
	})
	h.Controller.HandleStep(w, r, &handlers)
}
