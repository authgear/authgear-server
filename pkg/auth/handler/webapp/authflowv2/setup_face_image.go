package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebAuthflowSetupFaceImageHTML = template.RegisterHTML(
	"web/authflowv2/setup_face_image.html",
	handlerwebapp.Components...,
)

var AuthflowSetupFaceImageSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_b64_image": { "type": "string" }
		},
		"required": ["x_b64_image"]
	}
`)

func ConfigureAuthflowV2SetupFaceImageRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSetupFaceImage)
}

type AuthflowV2SetupFaceImageNavigator interface {
	NavigateSetupFaceImageSuccessPage(s *webapp.AuthflowScreen, r *http.Request, webSessionID string) (result *webapp.Result)
}

type AuthflowV2SetupFaceImageHandler struct {
	Controller    *handlerwebapp.AuthflowController
	Navigator     AuthflowV2SetupFaceImageNavigator
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      handlerwebapp.Renderer
}

func (h *AuthflowV2SetupFaceImageHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	branchViewModel := viewmodels.NewAuthflowBranchViewModel(screen)
	viewmodels.Embed(data, branchViewModel)

	return data, nil
}

func (h *AuthflowV2SetupFaceImageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers

	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowSetupFaceImageHTML, data)
		return nil
	})

	handlers.PostAction("submit", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowSetupFaceImageSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		index := *screen.Screen.TakenBranchIndex
		flowResponse := screen.StateTokenFlowResponse
		screenData := flowResponse.Action.Data.(declarative.CreateAuthenticatorData)
		option := screenData.Options[index]
		authentication := option.Authentication

		b64Image := r.Form.Get("x_b64_image")

		input := map[string]interface{}{
			"authentication": authentication,
			"b64_image":      b64Image,
		}

		result, err := h.Controller.AdvanceWithInput(r, s, screen, input, nil)
		if err != nil {
			return err
		}

		newScreen, err := h.Controller.DelayScreen(r, s, screen.Screen, result)
		if err != nil {
			return err
		}

		newResult := h.Navigator.NavigateSetupFaceImageSuccessPage(newScreen, r, s.ID)
		newResult.WriteResponse(w, r)
		return nil
	})
	h.Controller.HandleStep(w, r, &handlers)
}
