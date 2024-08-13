package authflowv2

import (
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebAuthflowSetupOOBOTPHTML = template.RegisterHTML(
	"web/authflowv2/setup_oob_otp.html",
	handlerwebapp.Components...,
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

func ConfigureAuthflowV2SetupOOBOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSetupOOBOTP)
}

type AuthflowSetupOOBOTPViewModel struct {
	OOBAuthenticatorType    model.AuthenticatorType
	Channel                 model.AuthenticatorOOBChannel
	IsBotProtectionRequired bool
}

type AuthflowV2SetupOOBOTPHandler struct {
	Controller    *handlerwebapp.AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      handlerwebapp.Renderer
}

func NewAuthflowSetupOOBOTPViewModel(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) AuthflowSetupOOBOTPViewModel {
	index := *screen.Screen.TakenBranchIndex
	screenData := screen.StateTokenFlowResponse.Action.Data.(declarative.IntentSignupFlowStepCreateAuthenticatorData)
	option := screenData.Options[index]

	var oobAuthenticatorType model.AuthenticatorType
	switch option.Authentication {
	case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
		oobAuthenticatorType = model.AuthenticatorTypeOOBEmail
	case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
		oobAuthenticatorType = model.AuthenticatorTypeOOBEmail
	case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
		oobAuthenticatorType = model.AuthenticatorTypeOOBSMS
	case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
		oobAuthenticatorType = model.AuthenticatorTypeOOBSMS
	default:
		panic(fmt.Errorf("unexpected authentication: %v", option.Authentication))
	}
	channel := screen.Screen.TakenChannel

	// Ignore error, bpRequired would be false
	bpRequired, _ := webapp.IsCreateAuthenticatorStepBotProtectionRequired(option.Authentication, screen.StateTokenFlowResponse)
	return AuthflowSetupOOBOTPViewModel{
		OOBAuthenticatorType:    oobAuthenticatorType,
		Channel:                 channel,
		IsBotProtectionRequired: bpRequired,
	}
}

func (h *AuthflowV2SetupOOBOTPHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenViewModel := NewAuthflowSetupOOBOTPViewModel(s, screen)
	viewmodels.Embed(data, screenViewModel)

	authentication := getTakenBranchSignupCreateAuthenticatorAuthentication(screen)

	branchFilter := func(branches []viewmodels.AuthflowBranch) []viewmodels.AuthflowBranch {
		filtered := []viewmodels.AuthflowBranch{}
		for _, branch := range branches {
			if branch.Authentication == authentication {
				// Hide oob otp branches of same type
				continue
			}
			filtered = append(filtered, branch)
		}
		return filtered
	}

	branchViewModel := viewmodels.NewAuthflowBranchViewModel(screen, branchFilter)
	viewmodels.Embed(data, branchViewModel)

	return data, nil
}

func (h *AuthflowV2SetupOOBOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowSetupOOBOTPHTML, data)
		return nil
	})
	handlers.PostAction("", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowSetupOOBOTPSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		index := *screen.Screen.TakenBranchIndex
		screenData := screen.StateTokenFlowResponse.Action.Data.(declarative.IntentSignupFlowStepCreateAuthenticatorData)
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

		result, err := h.Controller.AdvanceWithInput(r, s, screen, input, &handlerwebapp.AdvanceOptions{
			InheritTakenBranchState: true,
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
	h.Controller.HandleStep(w, r, &handlers)
}
