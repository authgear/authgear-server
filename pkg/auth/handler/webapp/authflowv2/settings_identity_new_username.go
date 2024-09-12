package authflowv2

import (
	"net/http"
	"net/url"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateSettingsIdentityNewUsernameTemplate = template.RegisterHTML(
	"web/authflowv2/settings_identity_new_username.html",
	handlerwebapp.SettingsComponents...,
)

var AuthflowV2SettingsIdentityNewUsernameSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_login_id_key": { "type": "string" },
			"x_login_id": { "type": "string" }
		},
		"required": ["x_login_id_key", "x_login_id"]
	}
`)

func ConfigureAuthflowV2SettingsIdentityNewUsername(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsIdentityNewUsername)
}

type AuthflowV2SettingsIdentityNewUsernameViewModel struct {
	LoginIDKey string
}

type AuthflowV2SettingsIdentityNewUsernameHandler struct {
	Database          *appdb.Handle
	AccountManagement *accountmanagement.Service
	ControllerFactory handlerwebapp.ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          handlerwebapp.Renderer
}

func (h *AuthflowV2SettingsIdentityNewUsernameHandler) GetData(w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, w)
	viewmodels.Embed(data, baseViewModel)

	loginIDKey := r.Form.Get("q_login_id_key")
	vm := AuthflowV2SettingsIdentityNewUsernameViewModel{
		LoginIDKey: loginIDKey,
	}
	viewmodels.Embed(data, vm)

	return data, nil
}

func (h *AuthflowV2SettingsIdentityNewUsernameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx()

	ctrl.Get(func() error {
		var data map[string]interface{}
		err := h.Database.WithTx(func() error {
			data, err = h.GetData(w, r)
			return err
		})
		if err != nil {
			return err
		}
		h.Renderer.RenderHTML(w, r, TemplateSettingsIdentityNewUsernameTemplate, data)
		return nil
	})

	ctrl.PostAction("", func() error {
		err := AuthflowV2SettingsIdentityNewUsernameSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}
		loginIDKey := r.Form.Get("x_login_id_key")
		loginID := r.Form.Get("x_login_id")
		resolvedSession := session.GetSession(r.Context())
		_, err = h.AccountManagement.AddUsername(resolvedSession, &accountmanagement.AddUsernameInput{
			LoginIDKey: loginIDKey,
			LoginID:    loginID,
		})
		if err != nil {
			return err
		}

		redirectURI, err := url.Parse(AuthflowV2RouteSettingsIdentityListUsername)
		if err != nil {
			return err
		}
		q := redirectURI.Query()
		q.Set("q_login_id_key", loginIDKey)
		redirectURI.RawQuery = q.Encode()
		result := webapp.Result{RedirectURI: redirectURI.String()}
		result.WriteResponse(w, r)

		return nil
	})
}
