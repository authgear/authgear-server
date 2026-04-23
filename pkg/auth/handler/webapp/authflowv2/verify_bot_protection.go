package authflowv2

import (
	"context"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebAuthflowVerifyBotProtectionHTML = template.RegisterHTML(
	"web/authflowv2/verify_bot_protection.html",
	handlerwebapp.Components...,
)

func ConfigureAuthflowV2VerifyBotProtectionRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteVerifyBotProtection)
}

type AuthflowV2VerifyBotProtectionHandler struct {
	Controller    *handlerwebapp.AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      handlerwebapp.Renderer
}

func (h *AuthflowV2VerifyBotProtectionHandler) GetData(w http.ResponseWriter, r *http.Request, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)
	return data, nil
}

func (h *AuthflowV2VerifyBotProtectionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowVerifyBotProtectionHTML, data)
		return nil
	})

	handlers.PostAction("", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := handlerwebapp.ValidateBotProtectionInput(ctx, r.Form)
		if err != nil {
			return err
		}

		result, err := h.advanceWithBotProtection(ctx, r, s, screen)
		if err != nil {
			return err
		}
		result.WriteResponse(w, r)
		return nil
	})

	h.Controller.HandleStep(r.Context(), w, r, &handlers)
}

func (h *AuthflowV2VerifyBotProtectionHandler) advanceWithBotProtection(
	ctx context.Context,
	r *http.Request,
	s *webapp.Session,
	screen *webapp.AuthflowScreenWithFlowResponse,
) (*webapp.Result, error) {
	index := *screen.Screen.TakenBranchIndex
	channel := screen.Screen.TakenChannel

	switch data := screen.StateTokenFlowResponse.Action.Data.(type) {
	case declarative.StepAuthenticateData:
		option := data.Options[index]

		input := map[string]interface{}{
			"authentication": option.Authentication,
			"index":          index,
		}
		if channel != "" {
			input["channel"] = channel
		}
		handlerwebapp.InsertBotProtection(r.Form, input)
		return h.Controller.AdvanceWithInput(ctx, r, s, screen, input, &handlerwebapp.AdvanceOptions{
			InheritTakenBranchState: true,
		})
	case declarative.AccountLinkingIdentifyData:
		option := data.Options[index]
		input := map[string]interface{}{
			"index": index,
		}
		if option.Identifcation == model.AuthenticationFlowIdentificationOAuth {
			redirectURI, err := h.Controller.GetAccountLinkingSSOCallbackURL(option.Alias, data)
			if err != nil {
				return nil, err
			}
			input["redirect_uri"] = redirectURI
		}
		handlerwebapp.InsertBotProtection(r.Form, input)
		return h.Controller.AdvanceWithInput(ctx, r, s, screen, input, nil)
	default:
		return nil, fmt.Errorf("unexpected data type: %T", screen.StateTokenFlowResponse.Action.Data)
	}
}
