package authflowv2

import (
	"context"
	"net/http"

	"net/url"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateSettingsIdentityViewUsernameTemplate = template.RegisterHTML(
	"web/authflowv2/settings_identity_view_username.html",
	handlerwebapp.SettingsComponents...,
)

var AuthflowV2SettingsIdentityDeleteUsernameSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_login_id_key": { "type": "string" },
			"x_identity_id": { "type": "string" }
		},
		"required": ["x_identity_id", "x_login_id_key"]
	}
`)

func ConfigureAuthflowV2SettingsIdentityViewUsername(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsIdentityViewUsername)
}

type AuthflowV2SettingsIdentityViewUsernameViewModel struct {
	LoginIDKey     string
	Identity       *identity.LoginID
	UpdateDisabled bool
	DeleteDisabled bool
}

type AuthflowV2SettingsIdentityViewUsernameHandler struct {
	Database          *appdb.Handle
	AccountManagement *accountmanagement.Service
	LoginIDConfig     *config.LoginIDConfig
	Identities        *identityservice.Service
	ControllerFactory handlerwebapp.ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          handlerwebapp.Renderer
}

func (h *AuthflowV2SettingsIdentityViewUsernameHandler) GetData(ctx context.Context, w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, w)
	viewmodels.Embed(data, baseViewModel)

	userID := session.GetUserID(ctx)
	loginID := r.Form.Get("q_login_id")
	usernameIdentity, err := h.Identities.LoginID.Get(ctx, *userID, loginID)
	if err != nil {
		return nil, err
	}

	updateDisabled := true
	deleteDisabled := true
	if loginIDConfig, ok := h.LoginIDConfig.GetKeyConfig(usernameIdentity.LoginIDKey); ok {
		updateDisabled = *loginIDConfig.UpdateDisabled
		deleteDisabled = *loginIDConfig.DeleteDisabled
	}

	identities, err := h.Identities.ListByUser(ctx, *userID)
	if err != nil {
		return nil, err
	}

	remaining := identity.ApplyFilters(
		identities,
		identity.KeepIdentifiable,
	)
	if len(remaining) == 1 {
		deleteDisabled = true
	}

	vm := AuthflowV2SettingsIdentityViewUsernameViewModel{
		LoginIDKey:     usernameIdentity.LoginIDKey,
		Identity:       usernameIdentity,
		UpdateDisabled: updateDisabled,
		DeleteDisabled: deleteDisabled,
	}
	viewmodels.Embed(data, vm)

	return data, nil
}

func (h *AuthflowV2SettingsIdentityViewUsernameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx(r.Context())

	ctrl.Get(func(ctx context.Context) error {
		var data map[string]interface{}
		err := h.Database.WithTx(ctx, func(ctx context.Context) error {
			data, err = h.GetData(ctx, w, r)
			return err
		})
		if err != nil {
			return err
		}
		h.Renderer.RenderHTML(w, r, TemplateSettingsIdentityViewUsernameTemplate, data)
		return nil
	})

	ctrl.PostAction("remove", func(ctx context.Context) error {
		err := AuthflowV2SettingsIdentityDeleteUsernameSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}
		identityID := r.Form.Get("x_identity_id")
		loginIDKey := r.Form.Get("x_login_key")

		resolvedSession := session.GetSession(ctx)
		_, err = h.AccountManagement.DeleteIdentityUsername(ctx, resolvedSession, &accountmanagement.DeleteIdentityUsernameInput{
			IdentityID: identityID,
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
