package authflowv2

import (
	"net/http"
	"sort"

	"github.com/authgear/authgear-server/pkg/api/model"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsIdentityListPhoneHTML = template.RegisterHTML(
	"web/authflowv2/settings_identity_list_phone.html",
	handlerwebapp.SettingsComponents...,
)

func ConfigureAuthflowV2SettingsIdentityListPhoneRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsIdentityListPhone)
}

type AuthflowV2SettingsIdentityListPhoneViewModel struct {
	LoginIDKey      string
	PhoneIdentities []*identity.LoginID
	Verifications   map[string][]verification.ClaimStatus
	CreateDisabled  bool
}

type AuthflowV2SettingsIdentityListPhoneHandler struct {
	Database          *appdb.Handle
	LoginIDConfig     *config.LoginIDConfig
	Identities        *identityservice.Service
	ControllerFactory handlerwebapp.ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Verification      handlerwebapp.SettingsVerificationService
	Renderer          handlerwebapp.Renderer
}

func (h *AuthflowV2SettingsIdentityListPhoneHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	loginIDKey := r.Form.Get("q_login_id_key")
	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	userID := session.GetUserID(r.Context())

	identities, err := h.Identities.LoginID.List(*userID)
	if err != nil {
		return nil, err
	}

	var phoneIdentities []*identity.LoginID
	var phoneInfos []*identity.Info
	for _, identity := range identities {
		if identity.LoginIDType == model.LoginIDKeyTypePhone {
			if loginIDKey == "" || identity.LoginIDKey == loginIDKey {
				phoneIdentities = append(phoneIdentities, identity)
				phoneInfos = append(phoneInfos, identity.ToInfo())
			}
		}
	}

	sort.Slice(phoneIdentities, func(i, j int) bool {
		return phoneIdentities[i].UpdatedAt.Before(phoneIdentities[j].UpdatedAt)
	})

	verifications, err := h.Verification.GetVerificationStatuses(phoneInfos)
	if err != nil {
		return nil, err
	}

	createDisabled := true
	if loginIDConfig, ok := h.LoginIDConfig.GetKeyConfig(loginIDKey); ok {
		createDisabled = *loginIDConfig.CreateDisabled
	}

	vm := AuthflowV2SettingsIdentityListPhoneViewModel{
		LoginIDKey:      loginIDKey,
		PhoneIdentities: phoneIdentities,
		Verifications:   verifications,
		CreateDisabled:  createDisabled,
	}
	viewmodels.Embed(data, vm)

	return data, nil
}

func (h *AuthflowV2SettingsIdentityListPhoneHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx()

	ctrl.Get(func() error {
		var data map[string]interface{}
		err := h.Database.WithTx(func() error {
			data, err = h.GetData(r, w)
			return err
		})
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsIdentityListPhoneHTML, data)
		return nil
	})
}
