package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebAuthflowForgotPasswordHTML = template.RegisterHTML(
	"web/authflow_forgot_password.html",
	components...,
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

func NewAuthFlowForgotPasswordViewModel(
	r *http.Request,
	initialScreen *webapp.AuthflowScreenWithFlowResponse,
	selectDestinationScreen *webapp.AuthflowScreenWithFlowResponse) AuthFlowForgotPasswordViewModel {

	loginIDInputType := ForgotPasswordLoginIDInputTypeEmail
	qLoginIDInputType := ForgotPasswordLoginIDInputType(r.Form.Get("q_login_id_input_type"))
	if qLoginIDInputType.IsValid() {
		loginIDInputType = qLoginIDInputType
	}

	loginID := r.Form.Get("q_login_id")

	data, ok := initialScreen.StateTokenFlowResponse.Action.Data.(declarative.IntentAccountRecoveryFlowStepIdentifyData)
	if !ok {
		panic("authflow webapp: unexpected data")
	}

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
	flowName := "default"
	var handlers AuthflowControllerHandlers

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

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowForgotPasswordHTML, data)
		return nil
	})

	handlers.PostAction("", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowForgotPasswordSchema.Validator().ValidateValue(FormToJSON(r.Form))
		if err != nil {
			return err
		}

		loginID := r.Form.Get("x_login_id")
		identification := r.Form.Get("x_login_id_type")

		result, err := h.Controller.AdvanceWithInput(r, s, screen, map[string]interface{}{
			"identification": identification,
			"login_id":       loginID,
			"index":          0,
		})

		if err != nil {
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

	h.Controller.HandleStartOfFlow(w, r, webapp.SessionOptions{}, authflow.FlowReference{
		Type: authflow.FlowTypeAccountRecovery,
		Name: flowName,
	}, &handlers, input)
}
