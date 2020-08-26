package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

const (
	// nolint: gosec
	TemplateItemTypeAuthUICreatePasswordHTML string = "auth_ui_create_password.html"
)

var TemplateAuthUICreatePasswordHTML = template.T{
	Type:                    TemplateItemTypeAuthUICreatePasswordHTML,
	IsHTML:                  true,
	TranslationTemplateType: TemplateItemTypeAuthUITranslationJSON,
	Defines:                 defines,
	ComponentTemplateTypes:  components,
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
	Alternatives      []CreateAuthenticatorAlternative
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

func (h *CreatePasswordHandler) GetData(r *http.Request, state *webapp.State, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, state.Error)
	identityInfo := graph.MustGetUserLastIdentity()

	passwordPolicyViewModel := viewmodels.NewPasswordPolicyViewModel(
		h.PasswordPolicy.PasswordPolicy(),
		state.Error,
	)

	alternatives, err := DeriveCreateAuthenticatorAlternatives(
		// Use current state ID because the current node should be NodeCreateAuthenticatorBegin.
		state.ID,
		graph,
		authn.AuthenticatorTypePassword,
	)
	if err != nil {
		return nil, err
	}

	createPasswordViewModel := CreatePasswordViewModel{
		IdentityDisplayID: identityInfo.DisplayID(),
		Alternatives:      alternatives,
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		err := h.Database.WithTx(func() error {
			state, graph, err := h.WebApp.Get(StateID(r))
			if err != nil {
				return err
			}

			data, err := h.GetData(r, state, graph)
			if err != nil {
				return err
			}

			h.Renderer.RenderHTML(w, r, TemplateItemTypeAuthUICreatePasswordHTML, data)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}

	if r.Method == "POST" {
		err := h.Database.WithTx(func() error {
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
