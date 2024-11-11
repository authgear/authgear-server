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

var TemplateWebAuthflowPromptCreatePasskeyHTML = template.RegisterHTML(
	"web/authflow_prompt_create_passkey.html",
	Components...,
)

func ConfigureAuthflowPromptCreatePasskeyRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(webapp.AuthflowRoutePromptCreatePasskey)
}

type AuthflowPromptCreatePasskeyViewModel struct {
	CreationOptionsJSON string
}

type AuthflowPromptCreatePasskeyHandler struct {
	Controller    *AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
}

func (h *AuthflowPromptCreatePasskeyHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenData := screen.StateTokenFlowResponse.Action.Data.(declarative.NodePromptCreatePasskeyData)
	creationOptionsJSONBytes, err := json.Marshal(screenData.CreationOptions)
	if err != nil {
		return nil, err
	}
	creationOptionsJSON := string(creationOptionsJSONBytes)

	screenViewModel := AuthflowPromptCreatePasskeyViewModel{
		CreationOptionsJSON: creationOptionsJSON,
	}
	viewmodels.Embed(data, screenViewModel)

	return data, nil
}

func (h *AuthflowPromptCreatePasskeyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers AuthflowControllerHandlers
	handlers.Get(func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowPromptCreatePasskeyHTML, data)
		return nil
	})
	handlers.PostAction("skip", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		input := map[string]interface{}{
			"skip": true,
		}

		result, err := h.Controller.AdvanceWithInput(ctx, r, s, screen, input, nil)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
	handlers.PostAction("", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		attestationResponseStr := r.Form.Get("x_attestation_response")

		var creationResponseJSON interface{}
		err := json.Unmarshal([]byte(attestationResponseStr), &creationResponseJSON)
		if err != nil {
			return err
		}

		input := map[string]interface{}{
			"creation_response": creationResponseJSON,
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
