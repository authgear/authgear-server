package authflowv2

import (
	"context"
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebAuthflowV2FillFormHTML = template.RegisterHTML(
	"web/authflowv2/fill_form.html",
	handlerwebapp.Components...,
)

func ConfigureAuthflowV2FillFormRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteFillForm)
}

type AuthflowV2FillFormViewModel struct {
	Form FormViewModel
	// FlowType is for the description.
	FlowType string
}

type AuthflowV2FillFormHandler struct {
	Controller    *handlerwebapp.AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      handlerwebapp.Renderer
}

func (h *AuthflowV2FillFormHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]any, error) {
	data := map[string]any{}

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenData := screen.StateTokenFlowResponse.Action.Data.(declarative.FillFormData)
	form, err := NewFormViewModel(screenData.Form, &baseViewModel)
	if err != nil {
		return nil, err
	}
	viewmodels.Embed(data, AuthflowV2FillFormViewModel{
		Form:     form,
		FlowType: string(screen.StateTokenFlowResponse.Type),
	})

	return data, nil
}

func (h *AuthflowV2FillFormHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowV2FillFormHTML, data)
		return nil
	})
	handlers.PostAction("submit", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		screenData := screen.StateTokenFlowResponse.Action.Data.(declarative.FillFormData)
		in := BuildFormFieldsInput(r.Form, screenData.Form)
		input := map[string]any{"fields": in.Fields}

		result, err := h.Controller.AdvanceWithInput(ctx, r, s, screen, input, nil)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
	h.Controller.HandleStep(r.Context(), w, r, &handlers)
}
