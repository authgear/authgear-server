package authflowv2

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
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

type ResetPasswordHandlerPasswordPolicy interface {
	PasswordPolicy() []password.Policy
	PasswordRules() string
}

type ResetPasswordHandlerResetPasswordService interface {
	ResetPasswordByEndUser(ctx context.Context, code string, newPassword string) error
}

type ResetPasswordHandlerDatabase interface {
	WithTx(ctx context.Context, do func(ctx context.Context) error) (err error)
}

type AuthflowV2ResetPasswordHandler struct {
	NonAuthflowControllerFactory handlerwebapp.ControllerFactory
	Controller                   *handlerwebapp.AuthflowController
	BaseViewModel                *viewmodels.BaseViewModeler
	Renderer                     handlerwebapp.Renderer
	AdminAPIResetPasswordPolicy  ResetPasswordHandlerPasswordPolicy
	ResetPassword                ResetPasswordHandlerResetPasswordService
	Database                     ResetPasswordHandlerDatabase
}

func (h *AuthflowV2ResetPasswordHandler) GetNonAuthflowData(w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModel(r, w)
	viewmodels.Embed(data, baseViewModel)

	passwordPolicyViewModel := viewmodels.NewPasswordPolicyViewModel(
		h.AdminAPIResetPasswordPolicy.PasswordPolicy(),
		h.AdminAPIResetPasswordPolicy.PasswordRules(),
		baseViewModel.RawError,
		viewmodels.GetDefaultPasswordPolicyViewModelOptions(),
	)

	viewmodels.Embed(data, passwordPolicyViewModel)

	return data, nil
}
func (h *AuthflowV2ResetPasswordHandler) GetAuthflowData(w http.ResponseWriter, r *http.Request, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
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

func (h *AuthflowV2ResetPasswordHandler) GetAuthflowErrorData(w http.ResponseWriter, r *http.Request, err error) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	baseViewModel.SetError(err)
	viewmodels.Embed(data, baseViewModel)

	return data, nil
}

func (h *AuthflowV2ResetPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if isURLFromAdminAPI(r) {
		h.serveHTTPNonAuthflow(w, r)
	} else {
		h.serveHTTPAuthflow(w, r)
	}
}

func (h *AuthflowV2ResetPasswordHandler) serveHTTPNonAuthflow(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.NonAuthflowControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer ctrl.ServeWithoutDBTx(r.Context())

	ctrl.Get(func(ctx context.Context) error {
		var data map[string]interface{}
		err = h.Database.WithTx(ctx, func(ctx context.Context) error {
			data, err = h.GetNonAuthflowData(w, r)
			if err != nil {
				return err
			}
			return nil
		})

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowResetPasswordHTML, data)
		return nil
	})

	ctrl.PostAction("", func(ctx context.Context) error {
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

		code := r.URL.Query().Get("code")
		err = h.Database.WithTx(ctx, func(ctx context.Context) error {
			return h.ResetPassword.ResetPasswordByEndUser(
				ctx,
				code,
				newPassword,
			)
		})
		if err != nil {
			return err
		}
		// reset success
		result := webapp.Result{RedirectURI: AuthflowV2RouteResetPasswordSuccess}
		result.WriteResponse(w, r)
		return nil
	})
}

func (h *AuthflowV2ResetPasswordHandler) serveHTTPAuthflow(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {

		data, err := h.GetAuthflowData(w, r, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowResetPasswordHTML, data)
		return nil
	})
	handlers.PostAction("", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
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

		result, err := h.Controller.AdvanceWithInput(ctx, r, s, screen, map[string]interface{}{
			"new_password": newPassword,
		}, nil)

		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	errorHandler := handlerwebapp.AuthflowControllerErrorHandler(func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) error {
		if !apierrors.IsKind(err, forgotpassword.PasswordResetFailed) {
			return err
		}
		data, err := h.GetAuthflowErrorData(w, r, err)
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

func isURLFromAdminAPI(r *http.Request) bool {
	return r.URL.Query().Get(otp.FromAdminAPIQueryKey) == "true"
}
