package authflowv2

import (
	"net/http"

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

var TemplateWebSettingsIdentityViewPhoneHTML = template.RegisterHTML(
	"web/authflowv2/settings_identity_view_phone.html",
	handlerwebapp.SettingsComponents...,
)

func ConfigureAuthflowV2SettingsIdentityViewPhoneRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsIdentityViewPhone)
}

type AuthflowV2SettingsIdentityViewPhoneViewModel struct {
	LoginIDKey     string
	PhoneIdentity  *identity.LoginID
	Verifications  map[string][]verification.ClaimStatus
	UpdateDisabled bool
	DeleteDisabled bool
}

type AuthflowV2SettingsIdentityViewPhoneHandler struct {
	Database          *appdb.Handle
	LoginIDConfig     *config.LoginIDConfig
	Identities        *identityservice.Service
	ControllerFactory handlerwebapp.ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Verification      handlerwebapp.SettingsVerificationService
	Renderer          handlerwebapp.Renderer
}

func (h *AuthflowV2SettingsIdentityViewPhoneHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	loginIDKey := r.Form.Get("q_login_id_key")
	loginID := r.Form.Get("q_login_id")

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	userID := session.GetUserID(r.Context())

	phoneIdentity, err := h.Identities.LoginID.Get(*userID, loginID)
	if err != nil {
		return nil, err
	}

	verifications, err := h.Verification.GetVerificationStatuses([]*identity.Info{phoneIdentity.ToInfo()})
	if err != nil {
		return nil, err
	}

	updateDisabled := true
	deleteDisabled := true
	if loginIDConfig, ok := h.LoginIDConfig.GetKeyConfig(loginIDKey); ok {
		updateDisabled = *loginIDConfig.UpdateDisabled
		deleteDisabled = *loginIDConfig.DeleteDisabled
	}

	vm := AuthflowV2SettingsIdentityViewPhoneViewModel{
		LoginIDKey:     loginIDKey,
		PhoneIdentity:  phoneIdentity,
		Verifications:  verifications,
		UpdateDisabled: updateDisabled,
		DeleteDisabled: deleteDisabled,
	}
	viewmodels.Embed(data, vm)

	return data, nil
}

func (h *AuthflowV2SettingsIdentityViewPhoneHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsIdentityViewPhoneHTML, data)
		return nil
	})
}
