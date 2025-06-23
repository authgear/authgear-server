package webapp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebAuthflowSetupOOBOTPHTML = template.RegisterHTML(
	"web/authflow_setup_oob_otp.html",
	Components...,
)

var AuthflowSetupOOBOTPSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_target": { "type": "string" }
		},
		"required": ["x_target"]
	}
`)

func ConfigureAuthflowSetupOOBOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(webapp.AuthflowRouteSetupOOBOTP)
}

type AuthflowSetupOOBOTPViewModel struct {
	OOBAuthenticatorType model.AuthenticatorType
	Channel              model.AuthenticatorOOBChannel
}

type AuthflowSetupOOBOTPHandler struct {
	Controller    *AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
}

func (h *AuthflowSetupOOBOTPHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	index := *screen.Screen.TakenBranchIndex
	screenData := screen.StateTokenFlowResponse.Action.Data.(declarative.CreateAuthenticatorData)
	option := screenData.Options[index]

	var oobAuthenticatorType model.AuthenticatorType
	switch option.Authentication {
	case model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
		oobAuthenticatorType = model.AuthenticatorTypeOOBEmail
	case model.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
		oobAuthenticatorType = model.AuthenticatorTypeOOBEmail
	case model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
		oobAuthenticatorType = model.AuthenticatorTypeOOBSMS
	case model.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
		oobAuthenticatorType = model.AuthenticatorTypeOOBSMS
	default:
		panic(fmt.Errorf("unexpected authentication: %v", option.Authentication))
	}
	channel := screen.Screen.TakenChannel
	screenViewModel := AuthflowSetupOOBOTPViewModel{
		OOBAuthenticatorType: oobAuthenticatorType,
		Channel:              channel,
	}
	viewmodels.Embed(data, screenViewModel)

	branchViewModel := viewmodels.NewAuthflowBranchViewModel(screen)
	viewmodels.Embed(data, branchViewModel)

	return data, nil
}

func (h *AuthflowSetupOOBOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers AuthflowControllerHandlers
	handlers.Get(func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowSetupOOBOTPHTML, data)
		return nil
	})
	handlers.PostAction("", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowSetupOOBOTPSchema.Validator().ValidateValue(ctx, FormToJSON(r.Form))
		if err != nil {
			return err
		}

		index := *screen.Screen.TakenBranchIndex
		screenData := screen.StateTokenFlowResponse.Action.Data.(declarative.CreateAuthenticatorData)
		option := screenData.Options[index]
		authentication := option.Authentication
		channel := screen.Screen.TakenChannel

		target := r.Form.Get("x_target")

		if channel == "" {
			channel = option.Channels[0]
		}

		input := map[string]interface{}{
			"authentication": authentication,
			"target":         target,
			"channel":        channel,
		}

		result, err := h.Controller.AdvanceWithInput(ctx, r, s, screen, input, &AdvanceOptions{
			InheritTakenBranchState: true,
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
	h.Controller.HandleStep(r.Context(), w, r, &handlers)
}
