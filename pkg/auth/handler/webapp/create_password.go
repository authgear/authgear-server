package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/password"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/template"
	"github.com/authgear/authgear-server/pkg/validation"
)

const (
	// nolint: gosec
	TemplateItemTypeAuthUICreatePasswordHTML config.TemplateItemType = "auth_ui_create_password.html"
)

var TemplateAuthUICreatePasswordHTML = template.Spec{
	Type:        TemplateItemTypeAuthUICreatePasswordHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
}

const CreatePasswordRequestSchema = "CreatePasswordRequestSchema"

var CreatePasswordSchema = validation.NewMultipartSchema("").
	Add(CreatePasswordRequestSchema, `
		{
			"type": "object",
			"properties": {
				"x_password": { "type": "string" }
			},
			"required": ["x_password"]
		}
	`).Instantiate()

func ConfigureCreatePasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/create_password")
}

type CreatePasswordViewModel struct {
	IdentityDisplayID string
}

type PasswordPolicy interface {
	PasswordPolicy() []password.Policy
}

type CreatePasswordHandler struct {
	Database       *db.Handle
	BaseViewModel  *viewmodels.BaseViewModeler
	Renderer       Renderer
	WebApp         WebAppService
	PasswordPolicy PasswordPolicy
}

func (h *CreatePasswordHandler) GetData(r *http.Request, state *webapp.State, graph *newinteraction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, state.Error)
	identityInfo := graph.MustGetUserLastIdentity()
	passwordPolicyViewModel := viewmodels.NewPasswordPolicyViewModel(
		h.PasswordPolicy.PasswordPolicy(),
		state.Error,
	)
	createPasswordViewModel := CreatePasswordViewModel{
		IdentityDisplayID: identityInfo.DisplayID(),
	}

	viewmodels.EmbedForm(data, r.Form)
	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, passwordPolicyViewModel)
	viewmodels.Embed(data, createPasswordViewModel)

	return data, nil
}

type CreatePasswordInput struct {
	Password string
}

var _ nodes.InputCreateAuthenticatorPassword = &CreatePasswordInput{}

// GetPassword implements InputCreateAuthenticatorPassword.
func (i *CreatePasswordInput) GetPassword() string {
	return i.Password
}

func (h *CreatePasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		h.Database.WithTx(func() error {
			state, graph, err := h.WebApp.Get(StateID(r))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			data, err := h.GetData(r, state, graph)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			h.Renderer.RenderHTML(w, r, TemplateItemTypeAuthUICreatePasswordHTML, data)
			return nil
		})
	}

	if r.Method == "POST" {
		h.Database.WithTx(func() error {
			result, err := h.WebApp.PostInput(StateID(r), func() (input interface{}, err error) {
				err = CreatePasswordSchema.PartValidator(CreatePasswordRequestSchema).ValidateValue(FormToJSON(r.Form))
				if err != nil {
					return
				}

				plainPassword := r.Form.Get("x_password")
				input = &CreatePasswordInput{
					Password: plainPassword,
				}
				return
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}
			result.WriteResponse(w, r)
			return nil
		})
	}
}
