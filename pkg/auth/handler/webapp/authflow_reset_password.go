package webapp

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	pwd "github.com/authgear/authgear-server/pkg/util/password"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebAuthflowResetPasswordHTML = template.RegisterHTML(
	"web/authflow_reset_password.html",
	Components...,
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

func ConfigureAuthflowResetPasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(webapp.AuthflowRouteResetPassword)
}

type AuthflowResetPasswordHandler struct {
	Controller    *AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      Renderer
}

func (h *AuthflowResetPasswordHandler) GetData(w http.ResponseWriter, r *http.Request, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
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

func (h *AuthflowResetPasswordHandler) GetErrorData(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	baseViewModel.SetError(err, errorutil.FormatTrackingID(ctx))
	viewmodels.Embed(data, baseViewModel)

	return data, nil
}

func (h *AuthflowResetPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers AuthflowControllerHandlers
	handlers.Get(func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {

		data, err := h.GetData(w, r, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowResetPasswordHTML, data)
		return nil
	})
	handlers.PostAction("", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowResetPasswordSchema.Validator().ValidateValue(ctx, FormToJSON(r.Form))
		if err != nil {
			return err
		}

		newPassword := r.Form.Get("x_password")
		confirmPassword := r.Form.Get("x_confirm_password")
		err = pwd.ConfirmPassword(newPassword, confirmPassword)
		if err != nil {
			return err
		}

		result, err := h.Controller.AdvanceWithInput(ctx, r, s, screen, map[string]interface{}{
			"new_password": newPassword,
		}, nil)

		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	errorHandler := AuthflowControllerErrorHandler(func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) error {
		if !apierrors.IsKind(err, forgotpassword.PasswordResetFailed) {
			return err
		}
		data, err := h.GetErrorData(ctx, w, r, err)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowResetPasswordHTML, data)
		return nil
	})

	code := r.URL.Query().Get("code")
	if code != "" {
		h.Controller.HandleResumeOfFlow(r.Context(), w, r, webapp.SessionOptions{}, &handlers, map[string]interface{}{
			"account_recovery_code": code,
		}, &errorHandler)
	} else {
		h.Controller.HandleStep(r.Context(), w, r, &handlers)
	}
}
