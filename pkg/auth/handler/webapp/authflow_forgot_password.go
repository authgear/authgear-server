package webapp

import (
	"context"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/whatsapp"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebAuthflowForgotPasswordHTML = template.RegisterHTML(
	"web/authflow_forgot_password.html",
	Components...,
)

var AuthflowForgotPasswordSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_login_id": { "type": "string" },
			"x_login_id_type": { "type": "string", "enum": ["phone", "email"] }
		},
		"required": ["x_login_id", "x_login_id_type"]
	}
`)

func ConfigureAuthflowForgotPasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(webapp.AuthflowRouteForgotPassword)
}

type ForgotPasswordLoginIDInputType string

const (
	ForgotPasswordLoginIDInputTypeEmail ForgotPasswordLoginIDInputType = "email"
	ForgotPasswordLoginIDInputTypePhone ForgotPasswordLoginIDInputType = "phone"
)

func (t ForgotPasswordLoginIDInputType) IsValid() bool {
	switch t {
	case ForgotPasswordLoginIDInputTypeEmail:
		fallthrough
	case ForgotPasswordLoginIDInputTypePhone:
		return true
	default:
		return false
	}
}

type AuthFlowForgotPasswordViewModel struct {
	LoginIDInputType    ForgotPasswordLoginIDInputType
	LoginID             string
	PhoneLoginIDEnabled bool
	EmailLoginIDEnabled bool
	LoginIDDisabled     bool
	OTPForm             string
}

func forgotPasswordGetInitialLoginIDInputType(data declarative.IntentAccountRecoveryFlowStepIdentifyData) ForgotPasswordLoginIDInputType {
	if len(data.Options) < 1 {
		return ForgotPasswordLoginIDInputTypeEmail
	}
	switch data.Options[0].Identification {
	case config.AuthenticationFlowAccountRecoveryIdentificationEmail:
		return ForgotPasswordLoginIDInputTypeEmail
	case config.AuthenticationFlowAccountRecoveryIdentificationPhone:
		return ForgotPasswordLoginIDInputTypePhone
	}
	return ForgotPasswordLoginIDInputTypeEmail
}

func NewAuthFlowForgotPasswordViewModel(
	r *http.Request,
	initialScreen *webapp.AuthflowScreenWithFlowResponse,
	selectDestinationScreen *webapp.AuthflowScreenWithFlowResponse) AuthFlowForgotPasswordViewModel {

	data, ok := initialScreen.StateTokenFlowResponse.Action.Data.(declarative.IntentAccountRecoveryFlowStepIdentifyData)
	if !ok {
		panic("authflow webapp: unexpected data")
	}

	loginIDInputType := forgotPasswordGetInitialLoginIDInputType(data)
	qLoginIDInputType := ForgotPasswordLoginIDInputType(r.Form.Get("q_login_id_input_type"))
	if qLoginIDInputType.IsValid() {
		loginIDInputType = qLoginIDInputType
	}

	loginID := r.Form.Get("q_login_id")

	phoneLoginIDEnabled := false
	emailLoginIDEnabled := false

	for _, opt := range data.Options {
		switch opt.Identification {
		case config.AuthenticationFlowAccountRecoveryIdentificationEmail:
			emailLoginIDEnabled = true
		case config.AuthenticationFlowAccountRecoveryIdentificationPhone:
			phoneLoginIDEnabled = true
		}
	}

	loginIDDisabled := !phoneLoginIDEnabled && !emailLoginIDEnabled

	otpForm := ""
	if selectDestinationScreen != nil {
		data2, ok := selectDestinationScreen.StateTokenFlowResponse.Action.
			Data.(declarative.IntentAccountRecoveryFlowStepSelectDestinationData)
		if ok && len(data2.Options) > 0 {
			otpForm = string(data2.Options[0].OTPForm)
		}
	}

	return AuthFlowForgotPasswordViewModel{
		LoginIDInputType:    loginIDInputType,
		LoginID:             loginID,
		PhoneLoginIDEnabled: phoneLoginIDEnabled,
		EmailLoginIDEnabled: emailLoginIDEnabled,
		LoginIDDisabled:     loginIDDisabled,
		OTPForm:             otpForm,
	}
}

type AuthflowForgotPasswordHandler struct {
	Controller    *AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
}

func (h *AuthflowForgotPasswordHandler) GetData(
	w http.ResponseWriter,
	r *http.Request,
	s *webapp.Session,
	initialScreen *webapp.AuthflowScreenWithFlowResponse,
	selectDestinationScreen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenViewModel := NewAuthFlowForgotPasswordViewModel(r, initialScreen, selectDestinationScreen)
	viewmodels.Embed(data, screenViewModel)

	return data, nil
}

func (h *AuthflowForgotPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers AuthflowControllerHandlers

	handlers.Get(func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {

		var screenIdentify *webapp.AuthflowScreenWithFlowResponse
		var screenSelectDestination *webapp.AuthflowScreenWithFlowResponse

		// screen can be identity or select_destination according to the query
		switch config.AuthenticationFlowStepType(screen.StateTokenFlowResponse.Action.Type) {
		case config.AuthenticationFlowStepTypeIdentify:
			screenIdentify = screen
		case config.AuthenticationFlowStepTypeSelectDestination:
			screenSelectDestination = screen
			var err error
			screenIdentify, err = h.Controller.GetScreen(ctx, s, screen.Screen.PreviousXStep)
			if err != nil {
				return err
			}
		}

		data, err := h.GetData(w, r, s, screenIdentify, screenSelectDestination)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowForgotPasswordHTML, data)
		return nil
	})

	handlers.PostAction("", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowForgotPasswordSchema.Validator().ValidateValue(FormToJSON(r.Form))
		if err != nil {
			return err
		}

		loginID := r.Form.Get("x_login_id")
		identification := r.Form.Get("x_login_id_type")

		inputs := h.makeInputs(screen, identification, loginID, 0)

		result, err := h.Controller.AdvanceWithInputs(ctx, r, s, screen, inputs, nil)
		if errors.Is(err, whatsapp.ErrInvalidWhatsappUser) {
			// The code failed to send because it is not a valid whatsapp user
			// Try again with sms if possible
			var fallbackErr error
			result, fallbackErr = h.fallbackToSMS(ctx, r, s, screen, identification, loginID)
			if errors.Is(fallbackErr, ErrNoFallbackAvailable) {
				return err
			} else if fallbackErr != nil {
				return fallbackErr
			}
		} else if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	identification := r.URL.Query().Get("q_login_id_input_type")
	loginID := r.URL.Query().Get("q_login_id")

	var input interface{} = nil

	if identification != "" && loginID != "" {
		input = map[string]interface{}{
			"identification": identification,
			"login_id":       loginID,
		}
	}

	h.Controller.HandleStartOfFlow(r.Context(), w, r, webapp.SessionOptions{}, authflow.FlowTypeAccountRecovery, &handlers, input)
}

func (h *AuthflowForgotPasswordHandler) fallbackToSMS(
	ctx context.Context,
	r *http.Request,
	s *webapp.Session,
	screen *webapp.AuthflowScreenWithFlowResponse,
	identification string,
	loginID string,
) (*webapp.Result, error) {
	options := []declarative.AccountRecoveryDestinationOption{}
	switch config.AuthenticationFlowStepType(screen.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		input := map[string]interface{}{
			"identification": identification,
			"login_id":       loginID,
		}
		output, err := h.Controller.FeedInputWithoutNavigate(ctx, screen.StateTokenFlowResponse.StateToken, input)
		if err != nil {
			return nil, err
		}
		if data, ok := output.FlowAction.Data.(declarative.IntentAccountRecoveryFlowStepSelectDestinationData); ok {
			options = data.Options
		}

	case config.AuthenticationFlowStepTypeSelectDestination:
		if data, ok := screen.StateTokenFlowResponse.Action.Data.(declarative.IntentAccountRecoveryFlowStepSelectDestinationData); ok {
			options = data.Options
		}
	}

	smsOptionIdx := -1
	for idx, option := range options {
		if option.Channel == declarative.AccountRecoveryChannelSMS {
			smsOptionIdx = idx
			break
		}
	}
	if smsOptionIdx == -1 {
		// No sms option is available, failing
		return nil, ErrNoFallbackAvailable
	}

	inputs := h.makeInputs(screen, identification, loginID, smsOptionIdx)
	return h.Controller.AdvanceWithInputs(ctx, r, s, screen, inputs, nil)
}

func (h *AuthflowForgotPasswordHandler) makeInputs(
	screen *webapp.AuthflowScreenWithFlowResponse,
	identification string,
	loginID string,
	optionIndex int) (inputs []map[string]interface{}) {
	// screen can be identity or select_destination depends on the query
	switch config.AuthenticationFlowStepType(screen.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		// We need data of both steps, so they must be two inputs
		inputs = []map[string]interface{}{
			{
				"identification": identification,
				"login_id":       loginID,
			},
			{
				"index": optionIndex,
			},
		}
	case config.AuthenticationFlowStepTypeSelectDestination:
		inputs = []map[string]interface{}{
			{
				"index": optionIndex,
			},
		}
	}
	return
}
