package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	pwd "github.com/authgear/authgear-server/pkg/util/password"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebCreatePasswordHTML = template.RegisterHTML(
	"web/create_password.html",
	Components...,
)

var CreatePasswordSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_password": { "type": "string" },
			"x_confirm_password": { "type": "string" },
			"x_stage": { "type": "string" }
		},
		"required": ["x_password", "x_confirm_password", "x_stage"]
	}
`)

func ConfigureCreatePasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/flows/create_password")
}

type CreatePasswordViewModel struct {
	IdentityDisplayID string
}

type PasswordPolicy interface {
	PasswordPolicy() []password.Policy
	PasswordRules() string
}

type CreatePasswordHandler struct {
	ControllerFactory         ControllerFactory
	BaseViewModel             *viewmodels.BaseViewModeler
	AlternativeStepsViewModel *viewmodels.AlternativeStepsViewModeler
	Renderer                  Renderer
	PasswordPolicy            PasswordPolicy
}

func (h *CreatePasswordHandler) GetData(r *http.Request, rw http.ResponseWriter, session *webapp.Session, graph *interaction.Graph) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)

	displayID := ""
	var node viewmodels.CreateAuthenticatorBeginNode
	if !graph.FindLastNode(&node) {
		panic("create_authenticator_begin: expected graph has node implementing CreateAuthenticatorBeginNode")
	}
	isPrimary := node.GetCreateAuthenticatorStage() == authn.AuthenticationStagePrimary
	if isPrimary {
		identityInfo := graph.MustGetUserLastIdentity()
		displayID = identityInfo.DisplayID()
	}

	passwordPolicyViewModel := viewmodels.NewPasswordPolicyViewModel(
		h.PasswordPolicy.PasswordPolicy(),
		h.PasswordPolicy.PasswordRules(),
		baseViewModel.RawError,
		&viewmodels.PasswordPolicyViewModelOptions{
			// Hide reuse password policy when creating new
			// password through web UI (sign up)
			IsNew: isPrimary,
		},
	)

	alternatives, err := h.AlternativeStepsViewModel.CreateAuthenticatorAlternatives(graph, webapp.SessionStepCreatePassword)
	if err != nil {
		return nil, err
	}

	createPasswordViewModel := CreatePasswordViewModel{
		IdentityDisplayID: displayID,
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, passwordPolicyViewModel)
	viewmodels.Embed(data, createPasswordViewModel)
	viewmodels.Embed(data, *alternatives)

	return data, nil
}

func (h *CreatePasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithDBTx()

	ctrl.Get(func() error {
		session, err := ctrl.InteractionSession()
		if err != nil {
			return err
		}

		graph, err := ctrl.InteractionGet()
		if err != nil {
			return err
		}

		data, err := h.GetData(r, w, session, graph)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebCreatePasswordHTML, data)
		return nil
	})

	ctrl.PostAction("", func() error {
		result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
			err = CreatePasswordSchema.Validator().ValidateValue(FormToJSON(r.Form))
			if err != nil {
				return
			}

			newPassword := r.Form.Get("x_password")
			confirmPassword := r.Form.Get("x_confirm_password")
			stage := r.Form.Get("x_stage")
			err = pwd.ConfirmPassword(newPassword, confirmPassword)
			if err != nil {
				return
			}

			input = &InputSetupPassword{
				Stage:    stage,
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

	handleAlternativeSteps(ctrl)
}
