package authflowv2

import (
	"net/http"
	"net/url"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebSettingsIdentityAddPhoneHTML = template.RegisterHTML(
	"web/authflowv2/settings_identity_add_phone.html",
	handlerwebapp.SettingsComponents...,
)

var AuthflowV2SettingsIdentityAddPhoneSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_login_id": { "type": "string" }
		},
		"required": ["x_login_id"]
	}
`)

func ConfigureAuthflowV2SettingsIdentityAddPhoneRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsIdentityAddPhone)
}

type AuthflowV2SettingsIdentityAddPhoneViewModel struct {
	LoginIDKey string
}

type AuthflowV2SettingsIdentityAddPhoneHandler struct {
	ControllerFactory   handlerwebapp.ControllerFactory
	BaseViewModel       *viewmodels.BaseViewModeler
	Renderer            handlerwebapp.Renderer
	AuthenticatorConfig *config.AuthenticatorConfig
	AccountManagement   accountmanagement.Service
}

func (h *AuthflowV2SettingsIdentityAddPhoneHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	loginIDKey := r.Form.Get("q_login_id_key")

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	vm := AuthflowV2SettingsIdentityAddPhoneViewModel{
		LoginIDKey: loginIDKey,
	}
	viewmodels.Embed(data, vm)

	return data, nil
}

func (h *AuthflowV2SettingsIdentityAddPhoneHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx()

	ctrl.Get(func() error {
		data, err := h.GetData(r, w)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsIdentityAddPhoneHTML, data)
		return nil
	})

	ctrl.PostAction("", func() error {
		loginIDKey := r.Form.Get("q_login_id_key")

		err := AuthflowV2SettingsIdentityAddPhoneSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		loginID := r.Form.Get("x_login_id")

		s := session.GetSession(r.Context())
		output, err := h.AccountManagement.StartAddIdentityPhone(s, &accountmanagement.StartAddIdentityPhoneInput{
			LoginID:    loginID,
			LoginIDKey: loginIDKey,
		})
		if err != nil {
			return err
		}

		var redirectURI *url.URL
		if output.NeedVerification {
			redirectURI, err = url.Parse(AuthflowV2RouteSettingsIdentityVerifyPhone)

			q := redirectURI.Query()
			q.Set("q_login_id_key", loginIDKey)
			q.Set("q_token", output.Token)

			redirectURI.RawQuery = q.Encode()
		} else {
			redirectURI, err = url.Parse(AuthflowV2RouteSettingsIdentityListPhone)

			q := redirectURI.Query()
			q.Set("q_login_id_key", loginIDKey)

			redirectURI.RawQuery = q.Encode()
		}
		if err != nil {
			return err
		}

		result := webapp.Result{RedirectURI: redirectURI.String()}
		result.WriteResponse(w, r)
		return nil
	})
}
