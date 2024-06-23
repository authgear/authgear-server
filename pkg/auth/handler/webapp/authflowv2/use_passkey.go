package authflowv2

import (
	"encoding/json"
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebAuthflowUsePasskeyHTML = template.RegisterHTML(
	"web/authflowv2/use_passkey.html",
	handlerwebapp.Components...,
)

func ConfigureAuthflowV2UsePasskeyRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteUsePasskey)
}

type AuthflowV2UsePasskeyViewModel struct {
	AutoExecute               bool
	PasskeyRequestOptionsJSON string
}

func NewAuthflowV2UsePasskeyViewModel(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (*AuthflowV2UsePasskeyViewModel, error) {
	index := *screen.Screen.TakenBranchIndex
	flowResponse := screen.BranchStateTokenFlowResponse
	data := flowResponse.Action.Data.(declarative.StepAuthenticateData)
	option := data.Options[index]

	requestOptionsJSONBytes, err := json.Marshal(option.RequestOptions)
	if err != nil {
		return nil, err
	}

	return &AuthflowV2UsePasskeyViewModel{
		AutoExecute:               true,
		PasskeyRequestOptionsJSON: string(requestOptionsJSONBytes),
	}, nil
}

type AuthflowV2UsePasskeyHandler struct {
	Controller                             *handlerwebapp.AuthflowController
	BaseViewModel                          *viewmodels.BaseViewModeler
	InlinePreviewAuthflowBranchViewModeler *viewmodels.InlinePreviewAuthflowBranchViewModeler
	Renderer                               handlerwebapp.Renderer
}

func (h *AuthflowV2UsePasskeyHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenViewModel, err := NewAuthflowV2UsePasskeyViewModel(s, screen)
	if err != nil {
		return nil, err
	}
	viewmodels.Embed(data, *screenViewModel)

	branchViewModel := viewmodels.NewAuthflowBranchViewModel(screen)
	viewmodels.Embed(data, branchViewModel)

	return data, nil
}

func (h *AuthflowV2UsePasskeyHandler) GetInlinePreviewData(w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForInlinePreviewAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenViewModel := AuthflowV2UsePasskeyViewModel{
		AutoExecute:               false,
		PasskeyRequestOptionsJSON: "{}",
	}
	viewmodels.Embed(data, screenViewModel)

	branchViewModel := h.InlinePreviewAuthflowBranchViewModeler.NewAuthflowBranchViewModelForInlinePreviewUsePasskey()
	viewmodels.Embed(data, branchViewModel)

	return data, nil
}

func (h *AuthflowV2UsePasskeyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowUsePasskeyHTML, data)
		return nil
	})
	handlers.PostAction("", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
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

		result, err := h.Controller.AdvanceWithInput(r, s, screen, input, nil)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
	handlers.InlinePreview(func(w http.ResponseWriter, r *http.Request) error {
		data, err := h.GetInlinePreviewData(w, r)
		if err != nil {
			return err
		}
		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowUsePasskeyHTML, data)
		return nil
	})

	h.Controller.HandleStep(w, r, &handlers)
}
