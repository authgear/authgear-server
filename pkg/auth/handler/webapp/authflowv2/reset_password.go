package authflowv2

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	pwd "github.com/authgear/authgear-server/pkg/util/password"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebAuthflowResetPasswordHTML = template.RegisterHTML(
	"web/authflowv2/reset_password.html",
	handlerwebapp.Components...,
)

var AuthflowResetPasswordSchema = validation.NewSimpleSchema(`
{
  "type": "object",
  "properties": {
    "x_password": { "type": "string" },
    "x_confirm_password": { "type": "string" }
  },
  "required": ["x_password", "x_confirm_password"]
}
`)

func ConfigureAuthflowV2ResetPasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteResetPassword)
}

type AuthflowV2ResetPasswordHandler struct {
	Controller    *handlerwebapp.AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      handlerwebapp.Renderer
}

func (h *AuthflowV2ResetPasswordHandler) GetData(w http.ResponseWriter, r *http.Request, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenData := screen.StateTokenFlowResponse.Action.Data.(declarative.NewPasswordData)

	passwordPolicyViewModel := viewmodels.NewPasswordPolicyViewModelFromAuthflow(
		screenData.PasswordPolicy,
		baseViewModel.RawError,
		&viewmodels.PasswordPolicyViewModelOptions{
			IsNew: false,
		},
	)

	viewmodels.Embed(data, passwordPolicyViewModel)

	return data, nil
}

func (h *AuthflowV2ResetPasswordHandler) GetErrorData(w http.ResponseWriter, r *http.Request, err error) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	baseViewModel.SetError(err)
	viewmodels.Embed(data, baseViewModel)

	return data, nil
}

func (h *AuthflowV2ResetPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {

		data, err := h.GetData(w, r, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowResetPasswordHTML, data)
		return nil
	})
	handlers.PostAction("", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowResetPasswordSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		newPassword := r.Form.Get("x_password")
		confirmPassword := r.Form.Get("x_confirm_password")
		err = pwd.ConfirmPassword(newPassword, confirmPassword)
		if err != nil {
			return err
		}

		result, err := h.Controller.AdvanceWithInput(r, s, screen, map[string]interface{}{
			"new_password": newPassword,
		}, nil)

		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	errorHandler := handlerwebapp.AuthflowControllerErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) error {
		if !apierrors.IsKind(err, forgotpassword.PasswordResetFailed) {
			return err
		}
		data, err := h.GetErrorData(w, r, err)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowResetPasswordHTML, data)
		return nil
	})

	code := r.URL.Query().Get("code")
	if code != "" {
		h.Controller.HandleResumeOfFlow(w, r, webapp.SessionOptions{}, &handlers, map[string]interface{}{
			"account_recovery_code": code,
		}, &errorHandler)
	} else {
		h.Controller.HandleStep(w, r, &handlers)
	}
}
