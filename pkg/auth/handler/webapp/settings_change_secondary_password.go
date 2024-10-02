package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	pwd "github.com/authgear/authgear-server/pkg/util/password"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var SettingsChangeSecondaryPasswordSchema = validation.NewSimpleSchema(`
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

func ConfigureSettingsChangeSecondaryPasswordRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "POST", "GET").WithPathPattern("/settings/mfa/change_secondary_password")
}

type SettingsChangeSecondaryPasswordHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
	PasswordPolicy    PasswordPolicy
}

func (h *SettingsChangeSecondaryPasswordHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	passwordPolicyViewModel := viewmodels.NewPasswordPolicyViewModel(
		h.PasswordPolicy.PasswordPolicy(),
		h.PasswordPolicy.PasswordRules(),
		baseViewModel.RawError,
		viewmodels.GetDefaultPasswordPolicyViewModelOptions(),
	)
	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, passwordPolicyViewModel)
	viewmodels.Embed(data, ChangePasswordViewModel{
		Force: false,
	})
	return data, nil
}

func (h *SettingsChangeSecondaryPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithDBTx()

	ctrl.Get(func() error {
		data, err := h.GetData(r, w)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebChangeSecondaryPasswordHTML, data)
		return nil
	})

	ctrl.PostAction("", func() error {
		userID := ctrl.RequireUserID()
		opts := webapp.SessionOptions{
			RedirectURI: "/settings",
		}
		intent := intents.NewIntentChangeSecondaryPassword(userID)

		result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
			err = SettingsChangeSecondaryPasswordSchema.Validator().ValidateValue(FormToJSON(r.Form))
			if err != nil {
				return
			}

			oldPassword := r.Form.Get("x_old_password")
			newPassword := r.Form.Get("x_new_password")
			confirmPassword := r.Form.Get("x_confirm_password")
			err = pwd.ConfirmPassword(newPassword, confirmPassword)
			if err != nil {
				return
			}

			input = &InputChangePassword{
				AuthenticationStage: authn.AuthenticationStageSecondary,
				OldPassword:         oldPassword,
				NewPassword:         newPassword,
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
