package authflowv2

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	authflowv2viewmodels "github.com/authgear/authgear-server/pkg/auth/handler/webapp/authflowv2/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"

	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebAuthflowLDAPLoginHTML = template.RegisterHTML(
	"web/authflowv2/ldap_login.html",
	handlerwebapp.Components...,
)

var AuthflowLDAPLoginSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_username": { "type": "string" },
			"x_password": { "type": "string" }
		},
		"required": ["x_username", "x_password"]
	}
`)

func ConfigureAuthflowV2LDAPLoginRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteLDAPLogin)
}

type AuthflowLDAPLoginViewModel struct {
	LDAPUsernameInputError *authflowv2viewmodels.InputError
	PasswordInputError     *authflowv2viewmodels.InputError
	HasUnknownError        bool
}

type AuthflowV2LDAPLoginHandler struct {
	Controller    *handlerwebapp.AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      handlerwebapp.Renderer
}

// nolint: gocognit
func NewAuthflowLDAPLoginViewModel(apiError *apierrors.APIError) AuthflowLDAPLoginViewModel {
	viewModel := AuthflowLDAPLoginViewModel{
		LDAPUsernameInputError: &authflowv2viewmodels.InputError{
			HasError:        false,
			HasErrorMessage: false,
		},
		PasswordInputError: &authflowv2viewmodels.InputError{
			HasError:        false,
			HasErrorMessage: false,
		},
	}
	if apiError != nil {
		switch apiError.Reason {
		case "InvalidCredentials":
			viewModel.LDAPUsernameInputError.HasError = true
			viewModel.PasswordInputError.HasError = true
			// Alert invalid credentials error
			viewModel.HasUnknownError = true
		case "ValidationFailed":
			for _, causes := range apiError.Info["causes"].([]interface{}) {
				if cause, ok := causes.(map[string]interface{}); ok {
					if kind, ok := cause["kind"].(string); ok {
						if kind == "required" {
							if details, ok := cause["details"].(map[string]interface{}); ok {
								if missing, ok := details["missing"].([]interface{}); ok {
									if viewmodels.SliceContains(missing, "x_username") {
										viewModel.LDAPUsernameInputError.HasError = true
										viewModel.LDAPUsernameInputError.HasErrorMessage = true
									} else if viewmodels.SliceContains(missing, "x_password") {
										viewModel.PasswordInputError.HasError = true
										viewModel.PasswordInputError.HasErrorMessage = true
									}
								}
							}
						}
					}
				}
			}
		}

		if !viewModel.LDAPUsernameInputError.HasError && !viewModel.PasswordInputError.HasError {
			// If it is not an error shown in inputs, it is an unknown error
			viewModel.HasUnknownError = true
		}
	}

	return viewModel
}

func (h *AuthflowV2LDAPLoginHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	authflowLDAPLoginViewModel := NewAuthflowLDAPLoginViewModel(baseViewModel.RawError)
	viewmodels.Embed(data, authflowLDAPLoginViewModel)

	return data, nil
}

func (h *AuthflowV2LDAPLoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowLDAPLoginHTML, data)
		return nil
	})
	handlers.PostAction("", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowLDAPLoginSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		serverName := r.URL.Query().Get("q_server_name")
		plainUsername := r.Form.Get("x_username")
		plainPassword := r.Form.Get("x_password")

		input := map[string]interface{}{
			"identification": "ldap",
			"server_name":    serverName,
			"username":       plainUsername,
			"password":       plainPassword,
		}

		result, err := h.Controller.ReplaceScreen(r, s, authflow.FlowTypeSignupLogin, input)

		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	h.Controller.HandleStep(w, r, &handlers)
}
