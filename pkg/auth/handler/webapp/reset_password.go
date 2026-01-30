package webapp

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	pwd "github.com/authgear/authgear-server/pkg/util/password"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebResetPasswordHTML = template.RegisterHTML(
	"web/reset_password.html",
	Components...,
)

var ResetPasswordSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"code": { "type": "string" },
			"x_password": { "type": "string" },
			"x_confirm_password": { "type": "string" }
		},
		"required": ["code", "x_password", "x_confirm_password"]
	}
`)

func ConfigureResetPasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/flows/reset_password")
}

type ResetPasswordService interface {
	VerifyCode(ctx context.Context, code string) (state *otp.State, err error)
}

type ResetPasswordHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
	PasswordPolicy    PasswordPolicy
	ResetPassword     ResetPasswordService
}

func (h *ResetPasswordHandler) GetData(ctx context.Context, r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	passwordPolicyViewModel := viewmodels.NewPasswordPolicyViewModel(
		h.PasswordPolicy.PasswordPolicy(),
		h.PasswordPolicy.PasswordRules(),
		baseViewModel.RawError,
		viewmodels.GetDefaultPasswordPolicyViewModelOptions(),
	)

	_, err := h.ResetPassword.VerifyCode(ctx, r.Form.Get("code"))
	if apierrors.IsKind(err, forgotpassword.PasswordResetFailed) {
		baseViewModel.SetError(ctx, err)
	} else if err != nil {
		// Ignore other errors (e.g. rate limit),
		// and let it (potentially) fail when submitting.
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, passwordPolicyViewModel)

	return data, nil
}

func (h *ResetPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithDBTx(r.Context())

	opts := webapp.SessionOptions{
		KeepAfterFinish: true,
	}
	intent := intents.NewIntentResetPassword()

	ctrl.Get(func(ctx context.Context) error {
		data, err := h.GetData(ctx, r, w)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebResetPasswordHTML, data)
		return nil
	})

	ctrl.PostAction("", func(ctx context.Context) error {
		result, err := ctrl.EntryPointPost(ctx, opts, intent, func() (input interface{}, err error) {
			err = ResetPasswordSchema.Validator().ValidateValue(ctx, FormToJSON(r.Form))
			if err != nil {
				return
			}

			code := r.Form.Get("code")
			newPassword := r.Form.Get("x_password")
			confirmPassword := r.Form.Get("x_confirm_password")
			err = pwd.ConfirmPassword(newPassword, confirmPassword)
			if err != nil {
				return
			}

			input = &InputResetPassword{
				Code:     code,
				Password: newPassword,
			}
			return
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
}
