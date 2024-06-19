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

var TemplateWebAuthflowPromptCreatePasskeyHTML = template.RegisterHTML(
	"web/authflowv2/prompt_create_passkey.html",
	handlerwebapp.Components...,
)

func ConfigureAuthflowV2PromptCreatePasskeyRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RoutePromptCreatePasskey)
}

type AuthflowV2PromptCreatePasskeyViewModel struct {
	CreationOptionsJSON string
}

type AuthflowV2PromptCreatePasskeyHandler struct {
	Controller    *handlerwebapp.AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      handlerwebapp.Renderer
}

func (h *AuthflowV2PromptCreatePasskeyHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenData := screen.StateTokenFlowResponse.Action.Data.(declarative.NodePromptCreatePasskeyData)
	creationOptionsJSONBytes, err := json.Marshal(screenData.CreationOptions)
	if err != nil {
		return nil, err
	}
	creationOptionsJSON := string(creationOptionsJSONBytes)

	screenViewModel := AuthflowV2PromptCreatePasskeyViewModel{
		CreationOptionsJSON: creationOptionsJSON,
	}
	viewmodels.Embed(data, screenViewModel)

	return data, nil
}

func (h *AuthflowV2PromptCreatePasskeyHandler) GetInlinePreviewData(w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForInlinePreviewAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenViewModel := AuthflowV2PromptCreatePasskeyViewModel{
		CreationOptionsJSON: "{}",
	}
	viewmodels.Embed(data, screenViewModel)

	return data, nil
}

func (h *AuthflowV2PromptCreatePasskeyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowPromptCreatePasskeyHTML, data)
		return nil
	})
	handlers.PostAction("skip", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		input := map[string]interface{}{
			"skip": true,
		}

		result, err := h.Controller.AdvanceWithInput(r, s, screen, input, nil)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
	handlers.PostAction("", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		attestationResponseStr := r.Form.Get("x_attestation_response")

		var creationResponseJSON interface{}
		err := json.Unmarshal([]byte(attestationResponseStr), &creationResponseJSON)
		if err != nil {
			return err
		}

		input := map[string]interface{}{
			"creation_response": creationResponseJSON,
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
		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowPromptCreatePasskeyHTML, data)
		return nil
	})

	if webapp.IsPreviewModeInline(r) {
		h.Controller.HandleInlinePreview(w, r, &handlers)
		return
	}
	h.Controller.HandleStep(w, r, &handlers)
}
