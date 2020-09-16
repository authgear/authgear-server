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
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

const (
	// nolint: gosec
	TemplateItemTypeAuthUIChangePasswordHTML string = "auth_ui_change_password.html"
)

var TemplateAuthUIChangePasswordHTML = template.Register(template.T{
	Type:                    TemplateItemTypeAuthUIChangePasswordHTML,
	IsHTML:                  true,
	TranslationTemplateType: TemplateItemTypeAuthUITranslationJSON,
	Defines:                 defines,
	ComponentTemplateTypes:  components,
})

const ChangePasswordRequestSchema = "ChangePasswordRequestSchema"

var ChangePasswordSchema = validation.NewMultipartSchema("").
	Add(ChangePasswordRequestSchema, `
		{
			"type": "object",
			"properties": {
				"x_password": { "type": "string" }
			},
			"required": ["x_password"]
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

func (h *ChangePasswordHandler) GetData(r *http.Request, state *webapp.State, graph *interaction.Graph) (map[string]interface{}, error) {
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
		OldStateID:  StateID(r),
		RedirectURI: "/settings",
		Intent:      intents.NewIntentChangePrimaryPassword(*userID),
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

			h.Renderer.RenderHTML(w, r, TemplateItemTypeAuthUIChangePasswordHTML, data)
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
