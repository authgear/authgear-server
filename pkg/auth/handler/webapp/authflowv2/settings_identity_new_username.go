package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateSettingsIdentityNewUsernameTemplate = template.RegisterHTML(
	"web/authflowv2/settings_identity_new_username.html",
	handlerwebapp.SettingsComponents...,
)

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
}
