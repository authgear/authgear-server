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
	pwd "github.com/authgear/authgear-server/pkg/util/password"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebCreatePasswordHTML = template.RegisterHTML(
	"web/create_password.html",
	components...,
)

const CreatePasswordRequestSchema = "CreatePasswordRequestSchema"

var CreatePasswordSchema = validation.NewMultipartSchema("").
	Add(CreatePasswordRequestSchema, `
		{
			"type": "object",
			"properties": {
				"x_password": { "type": "string" },
				"x_confirm_password": { "type": "string" }
			},
			"required": ["x_password", "x_confirm_password"]
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

	displayID := ""
	var node CreateAuthenticatorBeginNode
	if !graph.FindLastNode(&node) {
		panic("create_authenticator_begin: expected graph has node implementing CreateAuthenticatorBeginNode")
	}
	isPrimary := node.GetCreateAuthenticatorStage() == interaction.AuthenticationStagePrimary
	if isPrimary {
		identityInfo := graph.MustGetUserLastIdentity()
		displayID = identityInfo.DisplayID()
	}

	passwordPolicyViewModel := viewmodels.NewPasswordPolicyViewModel(
		h.PasswordPolicy.PasswordPolicy(),
		state.Error,
		&viewmodels.PasswordPolicyViewModelOptions{
			// Hide reuse password policy when creating new
			// password through web UI (sign up)
			IsNew: isPrimary,
		},
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
		IdentityDisplayID: displayID,
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

			h.Renderer.RenderHTML(w, r, TemplateWebCreatePasswordHTML, data)
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

				newPassword := r.Form.Get("x_password")
				confirmPassword := r.Form.Get("x_confirm_password")
				err = pwd.ConfirmPassword(newPassword, confirmPassword)
				if err != nil {
					return
				}

				input = &CreatePasswordInput{
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
