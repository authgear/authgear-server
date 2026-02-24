package authflowv2

import (
	"context"
	"net/http"

	"net/url"

	"github.com/authgear/authgear-server/pkg/api/model"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebSettingsIdentityViewPhoneHTML = template.RegisterHTML(
	"web/authflowv2/settings_identity_view_phone.html",
	handlerwebapp.SettingsComponents...,
)

var AuthflowV2SettingsUpdateIdentityVerificationPhoneSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_channel": { "type": "string" },
			"x_login_id": { "type": "string" },
			"x_identity_id": { "type": "string" }
		},
		"required": ["x_channel", "x_login_id", "x_identity_id"]
	}
`)

var AuthflowV2SettingsRemoveIdentityPhoneSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_identity_id": { "type": "string" }
		},
		"required": ["x_identity_id"]
	}
`)

func ConfigureAuthflowV2SettingsIdentityViewPhoneRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsIdentityViewPhone)
}

type AuthflowV2SettingsIdentityViewPhoneViewModel struct {
	LoginIDKey          string
	Channel             string
	PhoneIdentity       *identity.LoginID
	Verified            bool
	UpdateDisabled      bool
	DeleteDisabled      bool
	VerificationEnabled bool
}

type AuthflowV2SettingsIdentityViewPhoneHandler struct {
	Database            *appdb.Handle
	LoginIDConfig       *config.LoginIDConfig
	VerificationConfig  *config.VerificationConfig
	Identities          SettingsIdentityService
	ControllerFactory   handlerwebapp.ControllerFactory
	BaseViewModel       *viewmodels.BaseViewModeler
	Verification        SettingsVerificationService
	Renderer            handlerwebapp.Renderer
	AuthenticatorConfig *config.AuthenticatorConfig
	AccountManagement   accountmanagement.Service
}

func (h *AuthflowV2SettingsIdentityViewPhoneHandler) GetData(ctx context.Context, r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	loginIDKey := r.Form.Get("q_login_id_key")
	identityID := r.Form.Get("q_identity_id")

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	userID := session.GetUserID(ctx)

	channel := h.AuthenticatorConfig.OOB.SMS.PhoneOTPMode.GetDefaultChannel()

	phoneIdentity, err := h.Identities.GetWithUserID(ctx, *userID, identityID)
	if err != nil {
		return nil, err
	}

	verified, err := h.AccountManagement.CheckIdentityVerified(ctx, phoneIdentity)
	if err != nil {
		return nil, err
	}

	updateDisabled := true
	deleteDisabled := true
	if loginIDConfig, ok := h.LoginIDConfig.GetKeyConfig(loginIDKey); ok {
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

	vm := AuthflowV2SettingsIdentityViewPhoneViewModel{
		LoginIDKey:          loginIDKey,
		Channel:             string(channel),
		PhoneIdentity:       phoneIdentity.LoginID,
		Verified:            verified,
		UpdateDisabled:      updateDisabled,
		DeleteDisabled:      deleteDisabled,
		VerificationEnabled: *h.VerificationConfig.Claims.PhoneNumber.Enabled,
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

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsIdentityViewPhoneHTML, data)
		return nil
	})

	ctrl.PostAction("remove", func(ctx context.Context) error {
		err := AuthflowV2SettingsRemoveIdentityPhoneSchema.Validator().ValidateValue(ctx, handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		removeID := r.Form.Get("x_identity_id")

		_, err = h.AccountManagement.DeleteIdentityPhone(ctx, session.GetSession(ctx), &accountmanagement.DeleteIdentityPhoneInput{
			IdentityID: removeID,
		})
		if err != nil {
			return err
		}

		loginIDKey := r.Form.Get("q_login_id_key")
		redirectURI, err := url.Parse(AuthflowV2RouteSettingsIdentityListPhone)
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

	ctrl.PostAction("verify", func(ctx context.Context) error {
		loginIDKey := r.Form.Get("q_login_id_key")

		err := AuthflowV2SettingsUpdateIdentityVerificationPhoneSchema.Validator().ValidateValue(ctx, handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		channel := model.AuthenticatorOOBChannel(r.Form.Get("x_channel"))
		loginID := r.Form.Get("x_login_id")
		identityID := r.Form.Get("x_identity_id")

		s := session.GetSession(ctx)
		output, err := h.AccountManagement.StartUpdateIdentityPhone(ctx, s, &accountmanagement.StartUpdateIdentityPhoneInput{
			Channel:                 channel,
			LoginID:                 loginID,
			LoginIDKey:              loginIDKey,
			IdentityID:              identityID,
			IsVerificationRequested: true,
		})
		if err != nil {
			return err
		}

		redirectURI, err := url.Parse(AuthflowV2RouteSettingsIdentityVerifyPhone)

		q := redirectURI.Query()
		q.Set("q_login_id_key", loginIDKey)
		q.Set("q_token", output.Token)

		redirectURI.RawQuery = q.Encode()
		if err != nil {
			return err
		}

		result := webapp.Result{RedirectURI: redirectURI.String()}
		result.WriteResponse(w, r)
		return nil
	})
}
