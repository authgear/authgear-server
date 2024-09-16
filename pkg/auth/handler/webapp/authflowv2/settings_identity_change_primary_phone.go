package authflowv2

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsIdentityChangePrimaryPhoneHTML = template.RegisterHTML(
	"web/authflowv2/settings_identity_change_primary_phone.html",
	handlerwebapp.SettingsComponents...,
)

func ConfigureAuthflowV2SettingsIdentityChangePrimaryPhoneRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsIdentityChangePrimaryPhone)
}

type AuthflowV2SettingsIdentityChangePrimaryPhoneViewModel struct {
	LoginIDKey      string
	PhoneIdentities []*identity.LoginID
}

type AuthflowV2SettingsIdentityChangePrimaryPhoneHandler struct {
	Database                 *appdb.Handle
	ControllerFactory        handlerwebapp.ControllerFactory
	BaseViewModel            *viewmodels.BaseViewModeler
	SettingsProfileViewModel *viewmodels.SettingsProfileViewModeler
	Renderer                 handlerwebapp.Renderer
	Users                    handlerwebapp.SettingsProfileEditUserService
	StdAttrs                 handlerwebapp.SettingsProfileEditStdAttrsService
	Identities               *identityservice.Service
}

func (h *AuthflowV2SettingsIdentityChangePrimaryPhoneHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	loginIDKey := r.Form.Get("q_login_id_key")

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	userID := session.GetUserID(r.Context())

	identities, err := h.Identities.LoginID.List(*userID)
	if err != nil {
		return nil, err
	}

	settingsProfileViewModel, err := h.SettingsProfileViewModel.ViewModel(*userID)
	if err != nil {
		return nil, err
	}
	viewmodels.Embed(data, *settingsProfileViewModel)

	var phoneIdentities []*identity.LoginID
	for _, identity := range identities {
		if identity.LoginIDType == model.LoginIDKeyTypePhone {
			phoneIdentities = append(phoneIdentities, identity)
		}
	}

	vm := AuthflowV2SettingsIdentityListPhoneViewModel{
		LoginIDKey:      loginIDKey,
		PhoneIdentities: phoneIdentities,
	}
	viewmodels.Embed(data, vm)

	return data, nil
}

func (h *AuthflowV2SettingsIdentityChangePrimaryPhoneHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsIdentityChangePrimaryPhoneHTML, data)
		return nil
	})

	ctrl.PostAction("save", func() error {
		userID := *session.GetUserID(r.Context())
		m := handlerwebapp.JSONPointerFormToMap(r.Form)

		err := h.Database.WithTx(func() error {
			u, err := h.Users.GetRaw(userID)
			if err != nil {
				return err
			}

			attrs, err := stdattrs.T(u.StandardAttributes).MergedWithForm(m)
			if err != nil {
				return err
			}

			err = h.StdAttrs.UpdateStandardAttributes(config.RoleEndUser, userID, attrs)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return err
		}

		result := webapp.Result{RedirectURI: "/settings/identity/phone"}
		result.WriteResponse(w, r)
		return nil
	})
}
