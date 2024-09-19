package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	authflowv2viewmodels "github.com/authgear/authgear-server/pkg/auth/handler/webapp/authflowv2/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"

	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	pwd "github.com/authgear/authgear-server/pkg/util/password"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebSettingsMFACreatePasswordHTML = template.RegisterHTML(
	"web/authflowv2/settings_mfa_create_password.html",
	handlerwebapp.SettingsComponents...,
)

var AuthflowV2SettingsMFACreatePasswordSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_password": { "type": "string" },
			"x_confirm_password": { "type": "string" }
		},
		"required": ["x_password", "x_confirm_password"]
	}
`)

func ConfigureAuthflowV2SettingsMFACreatePassword(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsMFACreatePassword)
}

type AuthflowV2SettingsMFACreatePasswordHandler struct {
	ControllerFactory handlerwebapp.ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	SettingsViewModel *viewmodels.SettingsViewModeler
	PasswordPolicy    handlerwebapp.PasswordPolicy
	Renderer          handlerwebapp.Renderer

	AccountManagementService *accountmanagement.Service
}

func (h *AuthflowV2SettingsMFACreatePasswordHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	passwordPolicyViewModel := viewmodels.NewPasswordPolicyViewModel(
		h.PasswordPolicy.PasswordPolicy(),
		h.PasswordPolicy.PasswordRules(),
		baseViewModel.RawError,
		viewmodels.GetDefaultPasswordPolicyViewModelOptions(),
	)
	viewmodels.Embed(data, passwordPolicyViewModel)

	passwordInputErrorViewModel := authflowv2viewmodels.NewPasswordInputErrorViewModel(baseViewModel.RawError)
	viewmodels.Embed(data, passwordInputErrorViewModel)

	return data, nil
}

func (h *AuthflowV2SettingsMFACreatePasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx()

	ctrl.Get(func() error {
		data, err := h.GetData(r, w)
		if err != nil {
			return err
		}
		h.Renderer.RenderHTML(w, r, TemplateWebSettingsMFACreatePasswordHTML, data)
		return nil
	})

	ctrl.PostAction("", func() error {
		err := AuthflowV2SettingsMFACreatePasswordSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		newPassword := r.Form.Get("x_password")
		confirmPassword := r.Form.Get("x_confirm_password")

		err = pwd.ConfirmPassword(newPassword, confirmPassword)
		if err != nil {
			return err
		}

		s := session.GetSession(r.Context())
		err = h.AccountManagementService.CreateAdditionalPassword(s, accountmanagement.CreateAdditionalPasswordInput{
			PlainPassword: newPassword,
		})
		if err != nil {
			return err
		}

		result := webapp.Result{RedirectURI: AuthflowV2RouteSettingsMFA}
		result.WriteResponse(w, r)

		return nil
	})
}
