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

var TemplateWebSettingsIdentityListEmailHTML = template.RegisterHTML(
	"web/authflowv2/settings_identity_list_email.html",
	handlerwebapp.SettingsComponents...,
)

func ConfigureAuthflowV2SettingsIdentityListEmailRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsIdentityListEmail)
}

type AuthflowV2SettingsIdentityListEmailViewModelOAuthIdentity struct {
	Email        string
	ProviderType string
}

type AuthflowV2SettingsIdentityListEmailViewModel struct {
	LoginIDKey           string
	PrimaryEmail         string
	AllEmails            []string
	EmailIdentities      []*identity.LoginID
	OAuthEmailIdentities []*AuthflowV2SettingsIdentityListEmailViewModelOAuthIdentity
	Verifications        map[string][]verification.ClaimStatus
	CreateDisabled       bool
}

type AuthflowV2SettingsIdentityListEmailHandler struct {
	Database                 *appdb.Handle
	LoginIDConfig            *config.LoginIDConfig
	Identities               *identityservice.Service
	ControllerFactory        handlerwebapp.ControllerFactory
	BaseViewModel            *viewmodels.BaseViewModeler
	Verification             SettingsVerificationService
	SettingsProfileViewModel *viewmodels.SettingsProfileViewModeler
	Renderer                 handlerwebapp.Renderer
}

func (h *AuthflowV2SettingsIdentityListEmailHandler) GetData(ctx context.Context, r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	loginIDKey := r.Form.Get("q_login_id_key")

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	userID := session.GetUserID(ctx)

	emails := setutil.Set[string]{}

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
	var emailIdentities []*identity.LoginID = []*identity.LoginID{}
	var oauthEmailIdentities []*AuthflowV2SettingsIdentityListEmailViewModelOAuthIdentity = []*AuthflowV2SettingsIdentityListEmailViewModelOAuthIdentity{}
	var emailInfos []*identity.Info = []*identity.Info{}
	for _, identity := range loginIDIdentities {
		if identity.LoginIDType == model.LoginIDKeyTypeEmail {
			emails.Add(identity.LoginID)
			emailIdentities = append(emailIdentities, identity)
			emailInfos = append(emailInfos, identity.ToInfo())
			if identity.LoginID == settingsProfileViewModel.Email {
				primary = identity.LoginID
			}
		}
	}

	for _, identity := range oauthIdentities {
		email, ok := identity.Claims[stdattrs.Email].(string)
		if ok && email != "" {
			emails.Add(email)
			oauthEmailIdentities = append(oauthEmailIdentities,
				&AuthflowV2SettingsIdentityListEmailViewModelOAuthIdentity{
					Email:        email,
					ProviderType: identity.ProviderID.Type,
				},
			)
			if email == settingsProfileViewModel.Email {
				primary = email
			}
		}
	}

	sort.Slice(emailIdentities, func(i, j int) bool {
		return emailIdentities[i].UpdatedAt.Before(emailIdentities[j].UpdatedAt)
	})

	verifications, err := h.Verification.GetVerificationStatuses(ctx, emailInfos)
	if err != nil {
		return nil, err
	}

	createDisabled := true
	if loginIDConfig, ok := h.LoginIDConfig.GetKeyConfig(loginIDKey); ok {
		createDisabled = *loginIDConfig.CreateDisabled
	}

	vm := AuthflowV2SettingsIdentityListEmailViewModel{
		LoginIDKey:           loginIDKey,
		EmailIdentities:      emailIdentities,
		OAuthEmailIdentities: oauthEmailIdentities,
		Verifications:        verifications,
		CreateDisabled:       createDisabled,
		PrimaryEmail:         primary,
		AllEmails:            emails.Keys(),
	}
	viewmodels.Embed(data, vm)

	return data, nil
}

func (h *AuthflowV2SettingsIdentityListEmailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsIdentityListEmailHTML, data)
		return nil
	})
}
