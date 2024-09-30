package authflowv2

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	pwd "github.com/authgear/authgear-server/pkg/util/password"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebAuthflowResetPasswordHTML = template.RegisterHTML(
	"web/authflowv2/reset_password.html",
	handlerwebapp.Components...,
)

var AuthflowResetPasswordSchema = validation.NewSimpleSchema(`
{
  "type": "object",
  "properties": {
    "x_password": { "type": "string" },
    "x_confirm_password": { "type": "string" }
  },
  "required": ["x_password", "x_confirm_password"]
}
`)

func ConfigureAuthflowV2ResetPasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteResetPassword)
}

type AuthflowV2ResetPasswordHandler struct {
	Controller    *handlerwebapp.AuthflowController
	BaseViewModel *viewmodels.BaseViewModeler
	Renderer      handlerwebapp.Renderer
}

func (h *AuthflowV2ResetPasswordHandler) GetNonAuthflowData(w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModel(r, w)
	viewmodels.Embed(data, baseViewModel)

	// TODO: embed password policy view model without authflow

	return data, nil
}
func (h *AuthflowV2ResetPasswordHandler) GetAuthflowData(w http.ResponseWriter, r *http.Request, screen *webapp.AuthflowScreenWithFlowResponse) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	viewmodels.Embed(data, baseViewModel)

	screenData := screen.StateTokenFlowResponse.Action.Data.(declarative.NewPasswordData)

	passwordPolicyViewModel := viewmodels.NewPasswordPolicyViewModelFromAuthflow(
		screenData.PasswordPolicy,
		baseViewModel.RawError,
		&viewmodels.PasswordPolicyViewModelOptions{
			IsNew: false,
		},
	)

	viewmodels.Embed(data, passwordPolicyViewModel)

	return data, nil
}

func (h *AuthflowV2ResetPasswordHandler) GetAuthflowErrorData(w http.ResponseWriter, r *http.Request, err error) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	baseViewModel := h.BaseViewModel.ViewModelForAuthFlow(r, w)
	baseViewModel.SetError(err)
	viewmodels.Embed(data, baseViewModel)

	return data, nil
}

func (h *AuthflowV2ResetPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if isURLFromAdminAPI(r) {
		h.serveHTTPNonAuthflow(w, r)
	} else {
		h.serveHTTPAuthflow(w, r)
	}
}

func (h *AuthflowV2ResetPasswordHandler) serveHTTPNonAuthflow(w http.ResponseWriter, r *http.Request) {
	makeHTTPHandler := func(handler func(w http.ResponseWriter, r *http.Request) error) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := handler(w, r)
			if err != nil {
				if apierrors.IsAPIError(err) {
					// TODO: render error
					// renderError(w, r, err)
				} else {
					panic(err)
				}
			}
		})
	}

	getHandler := makeHTTPHandler(func(w http.ResponseWriter, r *http.Request) error {
		data, err := h.GetNonAuthflowData(w, r)
		if err != nil {
			return err
		}
		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowResetPasswordHTML, data)
		return nil
	})

	postHandler := makeHTTPHandler(func(w http.ResponseWriter, r *http.Request) error {
		err := AuthflowResetPasswordSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}
		newPassword := r.Form.Get("x_password")
		confirmPassword := r.Form.Get("x_confirm_password")
		err = pwd.ConfirmPassword(newPassword, confirmPassword)
		if err != nil {
			return err
		}

		// TODO: update password without authflow

		return nil
	})

	switch r.Method {
	case "GET":
		getHandler.ServeHTTP(w, r)
	case "POST":
		postHandler.ServeHTTP(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

}

func (h *AuthflowV2ResetPasswordHandler) serveHTTPAuthflow(w http.ResponseWriter, r *http.Request) {
	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.Get(func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {

		data, err := h.GetAuthflowData(w, r, screen)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowResetPasswordHTML, data)
		return nil
	})
	handlers.PostAction("", func(s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		err := AuthflowResetPasswordSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		newPassword := r.Form.Get("x_password")
		confirmPassword := r.Form.Get("x_confirm_password")
		err = pwd.ConfirmPassword(newPassword, confirmPassword)
		if err != nil {
			return err
		}

		result, err := h.Controller.AdvanceWithInput(r, s, screen, map[string]interface{}{
			"new_password": newPassword,
		}, nil)

		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	errorHandler := handlerwebapp.AuthflowControllerErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) error {
		if !apierrors.IsKind(err, forgotpassword.PasswordResetFailed) {
			return err
		}
		data, err := h.GetAuthflowErrorData(w, r, err)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebAuthflowResetPasswordHTML, data)
		return nil
	})

	code := r.URL.Query().Get("code")
	if code != "" {
		h.Controller.HandleResumeOfFlow(w, r, webapp.SessionOptions{}, &handlers, map[string]interface{}{
			"account_recovery_code": code,
		}, &errorHandler)
	} else {
		h.Controller.HandleStep(w, r, &handlers)
	}
}

var fromAdminAPIQueryKey = "x_from_admin_api"

func isURLFromAdminAPI(r *http.Request) bool {
	return r.URL.Query().Get(fromAdminAPIQueryKey) == "true"
}
