package webapp

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebAuthflowUsePasskeyHTML = template.RegisterHTML(
	"web/authflow_use_passkey.html",
	Components...,
)

func ConfigureAuthflowUsePasskeyRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(webapp.AuthflowRouteUsePasskey)
}

type AuthflowUsePasskeyViewModel struct {
	PasskeyRequestOptionsJSON string
}

func NewAuthflowUsePasskeyViewModel(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (*AuthflowUsePasskeyViewModel, error) {
	index := *screen.Screen.TakenBranchIndex
	flowResponse := screen.BranchStateTokenFlowResponse
	data := flowResponse.Action.Data.(declarative.StepAuthenticateData)
	option := data.Options[index]

	requestOptionsJSONBytes, err := json.Marshal(option.RequestOptions)
	if err != nil {
		return nil, err
	}

	return &AuthflowUsePasskeyViewModel{
		PasskeyRequestOptionsJSON: string(requestOptionsJSONBytes),
	}, nil
}

type AuthflowUsePasskeyHandler struct {
	Controller    *AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
}

func (h *AuthflowUsePasskeyHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenViewModel, err := NewAuthflowUsePasskeyViewModel(s, screen)
	if err != nil {
		return nil, err
	}
	viewmodels.Embed(data, *screenViewModel)

	branchViewModel := viewmodels.NewAuthflowBranchViewModel(screen)
	viewmodels.Embed(data, branchViewModel)

	return data, nil
}

func (h *AuthflowUsePasskeyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers AuthflowControllerHandlers
	handlers.Get(func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowUsePasskeyHTML, data)
		return nil
	})
	handlers.PostAction("", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		assertionResponseStr := r.Form.Get("x_assertion_response")

		var assertionResponseJSON interface{}
		err := json.Unmarshal([]byte(assertionResponseStr), &assertionResponseJSON)
		if err != nil {
			return err
		}

		index := *screen.Screen.TakenBranchIndex
		flowResponse := screen.BranchStateTokenFlowResponse
		data := flowResponse.Action.Data.(declarative.StepAuthenticateData)
		option := data.Options[index]

		input := map[string]interface{}{
			"authentication":     option.Authentication,
			"assertion_response": assertionResponseJSON,
		}

		result, err := h.Controller.AdvanceWithInput(ctx, r, s, screen, input, nil)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
	h.Controller.HandleStep(r.Context(), w, r, &handlers)
}
