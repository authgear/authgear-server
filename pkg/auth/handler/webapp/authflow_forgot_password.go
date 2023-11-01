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

type AuthFlowForgotPasswordViewModel struct {
	LoginIDInputType    string
	LoginID             string
	PhoneLoginIDEnabled bool
	EmailLoginIDEnabled bool
	LoginIDDisabled     bool
}

func NewAuthFlowForgotPasswordViewModel(r *http.Request, screen *webapp.AuthflowScreenWithFlowResponse) AuthFlowForgotPasswordViewModel {
	loginIDInputType := r.Form.Get("q_login_id_input_type")
	loginID := r.Form.Get("q_login_id")

	data, ok := screen.StateTokenFlowResponse.Action.Data.(declarative.IntentAccountRecoveryFlowStepIdentifyData)
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

	return AuthFlowForgotPasswordViewModel{
		LoginIDInputType:    loginIDInputType,
		LoginID:             loginID,
		PhoneLoginIDEnabled: phoneLoginIDEnabled,
		EmailLoginIDEnabled: emailLoginIDEnabled,
		LoginIDDisabled:     loginIDDisabled,
	}
}

type AuthflowForgotPasswordHandler struct {
	Controller    *AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
}

func (h *AuthflowForgotPasswordHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenViewModel := NewAuthFlowForgotPasswordViewModel(r, screen)
	viewmodels.Embed(data, screenViewModel)

	return data, nil
}

func (h *AuthflowForgotPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flowName := "default"
	var handlers AuthflowControllerHandlers
	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {

		data, err := h.GetData(w, r, s, screen)
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

	h.Controller.HandleStartOfFlow(w, r, webapp.SessionOptions{}, authflow.FlowReference{
		Type: authflow.FlowTypeAccountRecovery,
		Name: flowName,
	}, &handlers)
}
