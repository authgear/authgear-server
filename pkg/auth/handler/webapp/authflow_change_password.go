package webapp

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebAuthflowChangePasswordHTML = template.RegisterHTML(
	"web/authflow_change_password.html",
	Components...,
)

var AuthflowChangePasswordSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_new_password": { "type": "string" }
		},
		"required": ["x_new_password"]
	}
`)

func ConfigureAuthflowChangePasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(webapp.AuthflowRouteChangePassword)
}

type AuthflowChangePasswordHandler struct {
	Controller              *AuthflowController
	BaseViewModel           *viewmodels.BaseViewModeler
	ChangePasswordViewModel *viewmodels.ChangePasswordViewModeler
	Renderer                Renderer
}

func (h *AuthflowChangePasswordHandler) GetData(w http.ResponseWriter, r *http.Request, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenData := screen.StateTokenFlowResponse.Action.Data.(declarative.ForceChangePasswordData)

	passwordPolicyViewModel := viewmodels.NewPasswordPolicyViewModelFromAuthflow(
		screenData.PasswordPolicy,
		baseViewModel.RawError,
		&viewmodels.PasswordPolicyViewModelOptions{
			IsNew: false,
		},
	)
	changePasswordViewModel := h.ChangePasswordViewModel.NewWithAuthflow(screenData.ForceChangeReason)
	viewmodels.Embed(data, passwordPolicyViewModel)
	viewmodels.Embed(data, changePasswordViewModel)

	return data, nil
}

func (h *AuthflowChangePasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlers AuthflowControllerHandlers
	handlers.Get(func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		data, err := h.GetData(w, r, s, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowChangePasswordHTML, data)
		return nil
	})
	handlers.PostAction("", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowChangePasswordSchema.Validator().ValidateValue(FormToJSON(r.Form))
		if err != nil {
			return err
		}

		newPlainPassword := r.Form.Get("x_new_password")

		input := map[string]interface{}{
			"new_password": newPlainPassword,
		}

		result, err := h.Controller.AdvanceWithInput(ctx, r, s, screen, input, nil)
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
	h.Controller.HandleStep(r.Context(), w, r, &handlers)
}
