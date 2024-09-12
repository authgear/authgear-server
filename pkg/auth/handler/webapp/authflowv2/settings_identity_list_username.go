package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateSettingsIdentityListUsernameTemplate = template.RegisterHTML(
	"web/authflowv2/settings_identity_list_username.html",
	handlerwebapp.SettingsComponents...,
)

func ConfigureAuthflowV2SettingsIdentityListUsername(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsIdentityListUsername)
}

type AuthflowV2SettingsIdentityListUsernameViewModel struct {
	LoginIDKey         string
	UsernameIdentities []*identity.LoginID // Expect to be all username login id
	CreateDisabled     bool
}

type AuthflowV2SettingsIdentityListUsernameHandler struct {
	Database                 *appdb.Handle
	LoginIDConfig            *config.LoginIDConfig
	Identities               *identityservice.Service
	ControllerFactory        handlerwebapp.ControllerFactory
	BaseViewModel            *viewmodels.BaseViewModeler
	SettingsProfileViewModel *viewmodels.SettingsProfileViewModeler
	Renderer                 handlerwebapp.Renderer
}

func (h *AuthflowV2SettingsIdentityListUsernameHandler) GetData(w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	loginIDKey := r.Form.Get("q_login_id_key")
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, w)
	viewmodels.Embed(data, baseViewModel)

	userID := session.GetUserID(r.Context())

	identities, err := h.Identities.LoginID.List(*userID)
	if err != nil {
		return nil, err
	}

	var usernameIdentities []*identity.LoginID
	for _, identity := range identities {
		if loginIDKey == "" || identity.LoginIDKey == loginIDKey {
			usernameIdentities = append(usernameIdentities, identity)
		}
	}

	createDisabled := true
	if loginIDConfig, ok := h.LoginIDConfig.GetKeyConfig(loginIDKey); ok {
		createDisabled = *loginIDConfig.CreateDisabled
	}

	vm := AuthflowV2SettingsIdentityListUsernameViewModel{
		LoginIDKey:         loginIDKey,
		UsernameIdentities: usernameIdentities,
		CreateDisabled:     createDisabled,
	}
	viewmodels.Embed(data, vm)

	return data, nil
}

func (h *AuthflowV2SettingsIdentityListUsernameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		h.Renderer.RenderHTML(w, r, TemplateSettingsIdentityListUsernameTemplate, data)
		return nil
	})
}
