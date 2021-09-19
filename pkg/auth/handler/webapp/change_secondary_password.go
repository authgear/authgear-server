package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	pwd "github.com/authgear/authgear-server/pkg/util/password"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebChangeSecondaryPasswordHTML = template.RegisterHTML(
	"web/change_secondary_password.html",
	components...,
)

var ChangeSecondaryPasswordSchema = validation.NewSimpleSchema(`
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

var ForceChangeSecondaryPasswordSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_new_password": { "type": "string" },
			"x_confirm_password": { "type": "string" }
		},
		"required": ["x_new_password", "x_confirm_password"]
	}
`)

func ConfigureForceChangeSecondaryPasswordRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "POST", "GET").WithPathPattern("/change_secondary_password")
}

func ConfigureChangeSecondaryPasswordRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "POST", "GET").WithPathPattern("/settings/mfa/change_secondary_password")
}

type ChangeSecondaryPasswordHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
	PasswordPolicy    PasswordPolicy
}

func (h *ChangeSecondaryPasswordHandler) GetData(r *http.Request, rw http.ResponseWriter, maybeGraph *interaction.Graph) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	passwordPolicyViewModel := viewmodels.NewPasswordPolicyViewModel(
		h.PasswordPolicy.PasswordPolicy(),
		baseViewModel.RawError,
		viewmodels.GetDefaultPasswordPolicyViewModelOptions(),
	)
	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, passwordPolicyViewModel)

	force := false
	var node ForceChangePasswordNode
	if maybeGraph != nil && maybeGraph.FindLastNode(&node) {
		force = node.IsForceChangePassword()
	}
	viewmodels.Embed(data, ChangePasswordViewModel{
		Force: force,
	})
	return data, nil
}

func (h *ChangeSecondaryPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	maybeWebSession := webapp.GetSession(r.Context())

	ctrl.Get(func() error {
		var err error
		var graph *interaction.Graph
		if maybeWebSession != nil {
			graph, err = ctrl.InteractionGetWithSession(maybeWebSession)
			if err != nil {
				return err
			}
		}

		data, err := h.GetData(r, w, graph)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebChangeSecondaryPasswordHTML, data)
		return nil
	})

	ctrl.PostAction("", func() error {
		if maybeWebSession != nil {
			result, err := ctrl.InteractionPost(func() (input interface{}, err error) {
				err = ForceChangeSecondaryPasswordSchema.Validator().ValidateValue(FormToJSON(r.Form))
				if err != nil {
					return
				}

				newPassword := r.Form.Get("x_new_password")
				confirmPassword := r.Form.Get("x_confirm_password")
				err = pwd.ConfirmPassword(newPassword, confirmPassword)
				if err != nil {
					return
				}

				input = &InputChangePassword{
					AuthenticationStage: authn.AuthenticationStageSecondary,
					NewPassword:         newPassword,
				}
				return
			})
			if err != nil {
				return err
			}

			result.WriteResponse(w, r)
			return nil
		}

		userID := ctrl.RequireUserID()
		opts := webapp.SessionOptions{
			RedirectURI: "/settings",
		}
		intent := intents.NewIntentChangeSecondaryPassword(userID)

		result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
			err = ChangeSecondaryPasswordSchema.Validator().ValidateValue(FormToJSON(r.Form))
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
