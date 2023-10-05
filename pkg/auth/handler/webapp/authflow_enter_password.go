package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebAuthflowEnterPasswordHTML = template.RegisterHTML(
	"web/authflow_enter_password.html",
	components...,
)

func ConfigureAuthflowEnterPasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(webapp.AuthflowRouteEnterPassword)
}

type AuthflowEnterPasswordViewModel struct {
	AuthenticationStage string
	FlowType            string
}

type AuthflowEnterPasswordHandler struct {
	Controller    *AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
}

func NewAuthflowEnterPasswordViewModel(screen *webapp.AuthflowScreenWithFlowResponse) AuthflowEnterPasswordViewModel {
	flowType := screen.StateTokenFlowResponse.Type

	index := *screen.Screen.TakenBranchIndex
	flowResponse := screen.BranchStateTokenFlowResponse
	data := flowResponse.Action.Data.(declarative.IntentLoginFlowStepAuthenticateData)
	option := data.Options[index]
	authenticationStage := authn.AuthenticationStageFromAuthenticationMethod(option.Authentication)

	return AuthflowEnterPasswordViewModel{
		AuthenticationStage: string(authenticationStage),
		FlowType:            string(flowType),
	}
}

func (h *AuthflowEnterPasswordHandler) GetData(w http.ResponseWriter, r *http.Request, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModel(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenViewModel := NewAuthflowEnterPasswordViewModel(screen)
	viewmodels.Embed(data, screenViewModel)

	return data, nil
}

func (h *AuthflowEnterPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers AuthflowControllerHandlers
	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowEnterPasswordHTML, data)
		return nil
	})
	handlers.PostAction("", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		// FIXME(authflow): feed input.
		return nil
	})
	h.Controller.HandleStep(w, r, &handlers)
}
