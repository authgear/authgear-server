package authflowv2

import (
	"errors"
	"net/http"
	"net/url"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebAuthflowV2ForgotPasswordHTML = template.RegisterHTML(
	"web/authflowv2/forgot_password.html",
	handlerwebapp.Components...,
)

var AuthflowV2ForgotPasswordSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_login_id": { "type": "string" },
			"x_login_id_type": { "type": "string", "enum": ["phone", "email"] }
		},
		"required": ["x_login_id", "x_login_id_type"]
	}
`)

func ConfigureAuthflowV2ForgotPasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteForgotPassword)
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

type AuthFlowV2ForgotPasswordAlternativeType string

const (
	AuthFlowV2ForgotPasswordAlternativeTypeEmail AuthFlowV2ForgotPasswordAlternativeType = "email"
	AuthFlowV2ForgotPasswordAlternativeTypePhone AuthFlowV2ForgotPasswordAlternativeType = "phone"
)

type AuthFlowV2ForgotPasswordAlternative struct {
	AlternativeType AuthFlowV2ForgotPasswordAlternativeType
	Href            string
}

type AuthFlowV2ForgotPasswordViewModel struct {
	LoginIDInputType     ForgotPasswordLoginIDInputType
	LoginID              string
	PhoneLoginIDEnabled  bool
	EmailLoginIDEnabled  bool
	LoginIDDisabled      bool
	RequiresLoginIDInput bool
	Alternatives         []*AuthFlowV2ForgotPasswordAlternative
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

func deriveForgotPasswordAlternatives(
	r *http.Request,
	loginIDInputType ForgotPasswordLoginIDInputType,
	emailLoginIDEnabled bool,
	phoneLoginIDEnabled bool) []*AuthFlowV2ForgotPasswordAlternative {

	alternatives := []*AuthFlowV2ForgotPasswordAlternative{}

	if loginIDInputType != ForgotPasswordLoginIDInputTypeEmail && emailLoginIDEnabled {
		alternatives = append(alternatives, &AuthFlowV2ForgotPasswordAlternative{
			AlternativeType: AuthFlowV2ForgotPasswordAlternativeTypeEmail,
			Href: webapp.MakeURL(r.URL, "", url.Values{
				"q_login_id_input_type": []string{string(ForgotPasswordLoginIDInputTypeEmail)},
			}).String(),
		})
	}

	if loginIDInputType != ForgotPasswordLoginIDInputTypePhone && phoneLoginIDEnabled {
		alternatives = append(alternatives, &AuthFlowV2ForgotPasswordAlternative{
			AlternativeType: AuthFlowV2ForgotPasswordAlternativeTypePhone,
			Href: webapp.MakeURL(r.URL, "", url.Values{
				"q_login_id_input_type": []string{string(ForgotPasswordLoginIDInputTypePhone)},
			}).String(),
		})
	}

	return alternatives
}

func NewAuthFlowV2ForgotPasswordViewModel(
	r *http.Request,
	initialScreen *webapp.AuthflowScreenWithFlowResponse,
	selectDestinationScreen *webapp.AuthflowScreenWithFlowResponse) AuthFlowV2ForgotPasswordViewModel {

	requiresLoginIDInput := true

	data, ok := initialScreen.StateTokenFlowResponse.Action.Data.(declarative.IntentAccountRecoveryFlowStepIdentifyData)
	if !ok {
		panic("authflow webapp: unexpected data")
	}

	loginIDInputType := forgotPasswordGetInitialLoginIDInputType(data)
	qLoginIDInputType := ForgotPasswordLoginIDInputType(r.Form.Get("q_login_id_input_type"))
	if qLoginIDInputType.IsValid() {
		loginIDInputType = qLoginIDInputType
	}

	qLoginID := r.Form.Get("q_login_id")
	loginID := qLoginID

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

	if selectDestinationScreen != nil {
		requiresLoginIDInput = false
	}

	if qLoginID != "" && qLoginIDInputType.IsValid() {
		requiresLoginIDInput = false
	}
	alternatives := deriveForgotPasswordAlternatives(
		r,
		loginIDInputType,
		emailLoginIDEnabled,
		phoneLoginIDEnabled,
	)

	return AuthFlowV2ForgotPasswordViewModel{
		LoginIDInputType:     loginIDInputType,
		LoginID:              loginID,
		PhoneLoginIDEnabled:  phoneLoginIDEnabled,
		EmailLoginIDEnabled:  emailLoginIDEnabled,
		LoginIDDisabled:      loginIDDisabled,
		RequiresLoginIDInput: requiresLoginIDInput,
		Alternatives:         alternatives,
	}
}

type AuthflowV2ForgotPasswordHandler struct {
	Controller        *handlerwebapp.AuthflowController
	BaseViewModel     *viewmodels.BaseViewModeler
	AuthflowViewModel *viewmodels.AuthflowViewModeler
	Renderer          handlerwebapp.Renderer
}

func (h *AuthflowV2ForgotPasswordHandler) GetData(
	w http.ResponseWriter,
	r *http.Request,
	s *webapp.Session,
	initialScreen *webapp.AuthflowScreenWithFlowResponse,
	selectDestinationScreen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	// Put authflowViewModel first to avoid other fields of authflowViewModel override below
	authflowViewModel := h.AuthflowViewModel.NewWithAccountRecoveryAuthflow(initialScreen.StateTokenFlowResponse, r)
	viewmodels.Embed(data, authflowViewModel)

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenViewModel := NewAuthFlowV2ForgotPasswordViewModel(r, initialScreen, selectDestinationScreen)
	viewmodels.Embed(data, screenViewModel)

	return data, nil
}

func (h *AuthflowV2ForgotPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers

	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {

		var screenIdentify *webapp.AuthflowScreenWithFlowResponse
		var screenSelectDestination *webapp.AuthflowScreenWithFlowResponse

		// screen can be identity or select_destination according to the query
		switch config.AuthenticationFlowStepType(screen.StateTokenFlowResponse.Action.Type) {
		case config.AuthenticationFlowStepTypeIdentify:
			screenIdentify = screen
		case config.AuthenticationFlowStepTypeSelectDestination:
			screenSelectDestination = screen
			var err error
			screenIdentify, err = h.Controller.GetScreen(s, screen.Screen.PreviousXStep)
			if err != nil {
				return err
			}
		}

		data, err := h.GetData(w, r, s, screenIdentify, screenSelectDestination)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowV2ForgotPasswordHTML, data)
		return nil
	})

	handlers.PostAction("", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowV2ForgotPasswordSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		loginID := r.Form.Get("x_login_id")
		identification := r.Form.Get("x_login_id_type")

		inputs := h.makeInputs(screen, identification, loginID, 0)

		for _, input := range inputs {
			err = handlerwebapp.HandleAccountRecoveryIdentificationBotProtection(config.AuthenticationFlowAccountRecoveryIdentification(identification), screen.StateTokenFlowResponse, r.Form, input)
			if err != nil {
				return err
			}
		}

		result, err := h.Controller.AdvanceWithInputs(r, s, screen, inputs, nil)
		if errors.Is(err, otp.ErrInvalidWhatsappUser) {
			// The code failed to send because it is not a valid whatsapp user
			// Try again with sms if possible
			var fallbackErr error
			result, fallbackErr = h.fallbackToSMS(r, s, screen, identification, loginID)
			if errors.Is(fallbackErr, handlerwebapp.ErrNoFallbackAvailable) {
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

	h.Controller.HandleStartOfFlow(w, r, webapp.SessionOptions{}, authflow.FlowTypeAccountRecovery, &handlers, nil)
}

func (h *AuthflowV2ForgotPasswordHandler) fallbackToSMS(
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
		output, err := h.Controller.FeedInputWithoutNavigate(screen.StateTokenFlowResponse.StateToken, input)
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
		return nil, handlerwebapp.ErrNoFallbackAvailable
	}

	inputs := h.makeInputs(screen, identification, loginID, smsOptionIdx)
	return h.Controller.AdvanceWithInputs(r, s, screen, inputs, nil)
}

func (h *AuthflowV2ForgotPasswordHandler) makeInputs(
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
