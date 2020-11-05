package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	pwd "github.com/authgear/authgear-server/pkg/util/password"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebChangePasswordHTML = template.RegisterHTML(
	"web/change_password.html",
	components...,
)

const ChangePasswordRequestSchema = "ChangePasswordRequestSchema"

var ChangePasswordSchema = validation.NewMultipartSchema("").
	Add(ChangePasswordRequestSchema, `
		{
			"type": "object",
			"properties": {
				"x_old_password": { "type": "string" },
				"x_new_password": { "type": "string" },
				"x_confirm_password": { "type": "string" }
			},
			"required": ["x_old_password", "x_new_password", "x_confirm_password"]
		}
	`).Instantiate()

func ConfigureChangePasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/change_password")
}

type ChangePasswordHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
	PasswordPolicy    PasswordPolicy
}

func (h *ChangePasswordHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	passwordPolicyViewModel := viewmodels.NewPasswordPolicyViewModel(
		h.PasswordPolicy.PasswordPolicy(),
		baseViewModel.RawError,
		viewmodels.GetDefaultPasswordPolicyViewModelOptions(),
	)
	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, passwordPolicyViewModel)
	return data, nil
}

func (h *ChangePasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := ctrl.RequireUserID()
	opts := webapp.SessionOptions{
		RedirectURI: "/settings",
	}
	intent := intents.NewIntentChangePrimaryPassword(userID)

	ctrl.Get(func() error {
		data, err := h.GetData(r, w)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebChangePasswordHTML, data)
		return nil
	})

	ctrl.PostAction("", func() error {
		result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
			err = ChangePasswordSchema.PartValidator(ChangePasswordRequestSchema).ValidateValue(FormToJSON(r.Form))
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
				OldPassword: oldPassword,
				NewPassword: newPassword,
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
