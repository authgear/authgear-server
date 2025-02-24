package authflowv2

import (
	"context"
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/session"
	pwd "github.com/authgear/authgear-server/pkg/util/password"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebSettingsV2ChangePasswordHTML = template.RegisterHTML(
	"web/authflowv2/settings_change_password.html",
	handlerwebapp.SettingsComponents...,
)

var AuthflowV2SettingsChangePasswordSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_old_password": { "type": "string" },
			"x_new_password": { "type": "string" },
			"x_confirm_password": { "type": "string" }
		},
		"required": ["x_old_password", "x_new_password", "x_confirm_password"]
	}
`)

type AuthflowV2SettingsChangePasswordHandler struct {
	ControllerFactory        handlerwebapp.ControllerFactory
	BaseViewModel            *viewmodels.BaseViewModeler
	Renderer                 handlerwebapp.Renderer
	AccountManagementService *accountmanagement.Service
	PasswordPolicy           handlerwebapp.PasswordPolicy
}

func (h *AuthflowV2SettingsChangePasswordHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	// BaseViewModel
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	passwordPolicyViewModel := viewmodels.NewPasswordPolicyViewModel(
		h.PasswordPolicy.PasswordPolicy(),
		h.PasswordPolicy.PasswordRules(),
		baseViewModel.RawError,
		viewmodels.GetDefaultPasswordPolicyViewModelOptions(),
	)
	viewmodels.Embed(data, passwordPolicyViewModel)

	viewmodels.Embed(data, handlerwebapp.ChangePasswordViewModel{
		Force: false,
	})

	return data, nil
}

func (h *AuthflowV2SettingsChangePasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx(r.Context())

	ctrl.GetWithSettingsActionWebSession(r, func(ctx context.Context, _ *webapp.Session) error {
		data, err := h.GetData(r, w)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsV2ChangePasswordHTML, data)

		return nil
	})

	ctrl.PostActionWithSettingsActionWebSession("", r, func(ctx context.Context, webappSession *webapp.Session) error {
		err := AuthflowV2SettingsChangePasswordSchema.Validator().ValidateValue(ctx, handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		oldPassword := r.Form.Get("x_old_password")
		newPassword := r.Form.Get("x_new_password")
		confirmPassword := r.Form.Get("x_confirm_password")

		err = pwd.ConfirmPassword(newPassword, confirmPassword)
		if err != nil {
			return err
		}

		s := session.GetSession(ctx)

		input := &accountmanagement.ChangePrimaryPasswordInput{
			OldPassword: oldPassword,
			NewPassword: newPassword,
		}

		err = h.AccountManagementService.ChangePrimaryPassword(ctx, s, input)
		if err != nil {
			return err
		}

		if ctrl.IsInSettingsAction(s, webappSession) {
			settingsActionResult, err := ctrl.FinishSettingsActionWithResult(ctx, s, webappSession)
			if err != nil {
				return err
			}
			settingsActionResult.WriteResponse(w, r)
			return nil
		}

		result := webapp.Result{
			NavigationAction: webapp.NavigationActionRedirect,
			RedirectURI:      SettingsV2RouteSettings,
		}
		result.WriteResponse(w, r)
		return nil
	})

}
