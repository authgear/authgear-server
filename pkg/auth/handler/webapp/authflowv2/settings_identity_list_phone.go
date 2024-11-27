package authflowv2

import (
	"context"
	"net/http"

	"sort"

	"github.com/authgear/authgear-server/pkg/api/model"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/setutil"
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

type AuthflowV2SettingsIdentityListPhoneViewModelOAuthIdentity struct {
	Phone        string
	ProviderType string
}

type AuthflowV2SettingsIdentityListPhoneViewModel struct {
	LoginIDKey           string
	PrimaryPhone         string
	AllPhones            []string
	PhoneIdentities      []*identity.LoginID
	OAuthPhoneIdentities []*AuthflowV2SettingsIdentityListPhoneViewModelOAuthIdentity
	Verifications        map[string][]verification.ClaimStatus
	CreateDisabled       bool
}

type AuthflowV2SettingsIdentityListPhoneHandler struct {
	Database                 *appdb.Handle
	LoginIDConfig            *config.LoginIDConfig
	Identities               *identityservice.Service
	ControllerFactory        handlerwebapp.ControllerFactory
	BaseViewModel            *viewmodels.BaseViewModeler
	Verification             handlerwebapp.SettingsVerificationService
	SettingsProfileViewModel *viewmodels.SettingsProfileViewModeler
	Renderer                 handlerwebapp.Renderer
}

func (h *AuthflowV2SettingsIdentityListPhoneHandler) GetData(ctx context.Context, r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	loginIDKey := r.Form.Get("q_login_id_key")

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	userID := session.GetUserID(ctx)

	phones := setutil.Set[string]{}

	loginIDIdentities, err := h.Identities.LoginID.List(ctx, *userID)
	if err != nil {
		return nil, err
	}

	oauthIdentities, err := h.Identities.OAuth.List(ctx, *userID)
	if err != nil {
		return nil, err
	}

	settingsProfileViewModel, err := h.SettingsProfileViewModel.ViewModel(ctx, *userID)
	if err != nil {
		return nil, err
	}

	var primary string
	var phoneIdentities []*identity.LoginID = []*identity.LoginID{}
	oauthPhoneIdentities := []*AuthflowV2SettingsIdentityListPhoneViewModelOAuthIdentity{}
	var phoneInfos []*identity.Info = []*identity.Info{}
	for _, identity := range loginIDIdentities {
		if identity.LoginIDType == model.LoginIDKeyTypePhone {
			phones.Add(identity.LoginID)
			phoneIdentities = append(phoneIdentities, identity)
			phoneInfos = append(phoneInfos, identity.ToInfo())
			if identity.LoginID == settingsProfileViewModel.PhoneNumber {
				primary = identity.LoginID
			}
		}
	}

	for _, identity := range oauthIdentities {
		phone, ok := identity.Claims[stdattrs.PhoneNumber].(string)
		if ok && phone != "" {
			phones.Add(phone)
			oauthPhoneIdentities = append(oauthPhoneIdentities,
				&AuthflowV2SettingsIdentityListPhoneViewModelOAuthIdentity{
					Phone:        phone,
					ProviderType: identity.ProviderID.Type,
				},
			)
			if phone == settingsProfileViewModel.PhoneNumber {
				primary = phone
			}
		}
	}

	sort.Slice(phoneIdentities, func(i, j int) bool {
		return phoneIdentities[i].UpdatedAt.Before(phoneIdentities[j].UpdatedAt)
	})

	verifications, err := h.Verification.GetVerificationStatuses(ctx, phoneInfos)
	if err != nil {
		return nil, err
	}

	createDisabled := true
	if loginIDConfig, ok := h.LoginIDConfig.GetKeyConfig(loginIDKey); ok {
		createDisabled = *loginIDConfig.CreateDisabled
	}

	vm := AuthflowV2SettingsIdentityListPhoneViewModel{
		LoginIDKey:           loginIDKey,
		PhoneIdentities:      phoneIdentities,
		OAuthPhoneIdentities: oauthPhoneIdentities,
		Verifications:        verifications,
		CreateDisabled:       createDisabled,
		PrimaryPhone:         primary,
		AllPhones:            phones.Keys(),
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
	defer ctrl.ServeWithoutDBTx(r.Context())

	ctrl.Get(func(ctx context.Context) error {
		var data map[string]interface{}
		err := h.Database.WithTx(ctx, func(ctx context.Context) error {
			data, err = h.GetData(ctx, r, w)
			return err
		})
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsIdentityListPhoneHTML, data)
		return nil
	})
}
