package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	pwd "github.com/authgear/authgear-server/pkg/util/password"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebChangeSecondaryPasswordHTML = template.RegisterHTML(
	"web/change_secondary_password.html",
	components...,
)

const ChangeSecondaryPasswordRequestSchema = "ChangeSecondaryPasswordRequestSchema"

var ChangeSecondaryPasswordSchema = validation.NewMultipartSchema("").
	Add(ChangeSecondaryPasswordRequestSchema, `
		{
			"type": "object",
			"properties": {
				"x_password": { "type": "string" },
				"x_confirm_password": { "type": "string" }
			},
			"required": ["x_password", "x_confirm_password"]
		}
	`).Instantiate()

func ConfigureChangeSecondaryPasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/change_secondary_password")
}

type ChangeSecondaryPasswordHandler struct {
	Database       *db.Handle
	BaseViewModel  *viewmodels.BaseViewModeler
	Renderer       Renderer
	WebApp         WebAppService
	PasswordPolicy PasswordPolicy
}

func (h *ChangeSecondaryPasswordHandler) GetData(r *http.Request, state *webapp.State, graph *interaction.Graph) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	var anyError interface{}
	if state != nil {
		anyError = state.Error
	}
	baseViewModel := h.BaseViewModel.ViewModel(r, anyError)
	passwordPolicyViewModel := viewmodels.NewPasswordPolicyViewModel(
		h.PasswordPolicy.PasswordPolicy(),
		state.Error,
	)
	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, passwordPolicyViewModel)
	return data, nil
}

func (h *ChangeSecondaryPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := session.GetUserID(r.Context())
	intent := &webapp.Intent{
		RedirectURI: "/settings",
		Intent:      intents.NewIntentChangeSecondaryPassword(*userID),
	}

	if r.Method == "GET" {
		err := h.Database.WithTx(func() error {
			state, graph, err := h.WebApp.GetIntent(intent)
			if err != nil {
				return err
			}

			data, err := h.GetData(r, state, graph)
			if err != nil {
				return err
			}

			h.Renderer.RenderHTML(w, r, TemplateWebChangeSecondaryPasswordHTML, data)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}

	if r.Method == "POST" {
		err := h.Database.WithTx(func() error {
			result, err := h.WebApp.PostIntent(intent, func() (input interface{}, err error) {
				err = ChangeSecondaryPasswordSchema.PartValidator(ChangeSecondaryPasswordRequestSchema).ValidateValue(FormToJSON(r.Form))
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
