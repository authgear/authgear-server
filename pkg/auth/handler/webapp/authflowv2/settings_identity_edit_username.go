package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateSettingsIdentityEditUsernameTemplate = template.RegisterHTML(
	"web/authflowv2/settings_identity_edit_username.html",
	handlerwebapp.SettingsComponents...,
)

func ConfigureAuthflowV2SettingsIdentityEditUsername(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsIdentityEditUsername)
}

type AuthflowV2SettingsIdentityEditUsernameViewModel struct {
	Identity *identity.LoginID
}

type AuthflowV2SettingsIdentityEditUsernameHandler struct {
	Database          *appdb.Handle
	Identities        *identityservice.Service
	ControllerFactory handlerwebapp.ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          handlerwebapp.Renderer
}

func (h *AuthflowV2SettingsIdentityEditUsernameHandler) GetData(w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, w)
	viewmodels.Embed(data, baseViewModel)

	userID := session.GetUserID(r.Context())
	loginID := r.Form.Get("q_login_id")
	usernameIdentity, err := h.Identities.LoginID.Get(*userID, loginID)
	if err != nil {
		return nil, err
	}

	vm := AuthflowV2SettingsIdentityEditUsernameViewModel{
		Identity: usernameIdentity,
	}
	viewmodels.Embed(data, vm)

	return data, nil
}

func (h *AuthflowV2SettingsIdentityEditUsernameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		h.Renderer.RenderHTML(w, r, TemplateSettingsIdentityEditUsernameTemplate, data)
		return nil
	})
}
