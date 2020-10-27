package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/session"
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
				"x_password": { "type": "string" },
				"x_confirm_password": { "type": "string" }
			},
			"required": ["x_password", "x_confirm_password"]
		}
	`).Instantiate()

func ConfigureChangePasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/change_password")
}

type ChangePasswordHandler struct {
	Database       *db.Handle
	BaseViewModel  *viewmodels.BaseViewModeler
	Renderer       Renderer
	WebApp         WebAppService
	PasswordPolicy PasswordPolicy
}

func (h *ChangePasswordHandler) GetData(r *http.Request, state *webapp.State) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	var anyError interface{}
	if state != nil {
		anyError = state.Error
	}
	baseViewModel := h.BaseViewModel.ViewModel(r, anyError)
	passwordPolicyViewModel := viewmodels.NewPasswordPolicyViewModel(
		h.PasswordPolicy.PasswordPolicy(),
		anyError,
	)
	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, passwordPolicyViewModel)
	return data, nil
}

type ChangePasswordInput struct {
	Password string
}

// GetNewPassword implements InputChangePassword.
func (i *ChangePasswordInput) GetNewPassword() string {
	return i.Password
}

func (h *ChangePasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := session.GetUserID(r.Context())
	intent := &webapp.Intent{
		RedirectURI: "/settings",
		OldStateID:  StateID(r),
		Intent:      intents.NewIntentChangePrimaryPassword(*userID),
	}

	if r.Method == "GET" {
		err := h.Database.WithTx(func() error {
			state, err := h.WebApp.GetState(StateID(r))
			if err != nil {
				return err
			}

			data, err := h.GetData(r, state)
			if err != nil {
				return err
			}

			h.Renderer.RenderHTML(w, r, TemplateWebChangePasswordHTML, data)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}

	if r.Method == "POST" {
		err := h.Database.WithTx(func() error {
			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				err = ChangePasswordSchema.PartValidator(ChangePasswordRequestSchema).ValidateValue(FormToJSON(r.Form))
				if err != nil {
					return
				}

				newPassword := r.Form.Get("x_password")
				confirmPassword := r.Form.Get("x_confirm_password")
				err = pwd.ConfirmPassword(newPassword, confirmPassword)
				if err != nil {
					return
				}

				input = &ChangePasswordInput{
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
		if err != nil {
			panic(err)
		}
	}
}
