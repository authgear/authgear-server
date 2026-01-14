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
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebSettingsIdentityViewEmailHTML = template.RegisterHTML(
	"web/authflowv2/settings_identity_view_email.html",
	handlerwebapp.SettingsComponents...,
)

var AuthflowV2SettingsIdentityUpdateVerificationEmailSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_login_id": { "type": "string" },
			"x_identity_id": { "type": "string" }
		},
		"required": ["x_login_id", "x_identity_id"]
	}
`)

var AuthflowV2SettingsRemoveIdentityEmailSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_identity_id": { "type": "string" }
		},
		"required": ["x_identity_id"]
	}
`)

func ConfigureAuthflowV2SettingsIdentityViewEmailRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsIdentityViewEmail)
}

type AuthflowV2SettingsIdentityViewEmailViewModel struct {
	LoginIDKey          string
	EmailIdentity       *identity.LoginID
	Verified            bool
	Verifications       map[string][]verification.ClaimStatus
	UpdateDisabled      bool
	DeleteDisabled      bool
	VerificationEnabled bool
}

type AuthflowV2SettingsIdentityViewEmailHandler struct {
	Database           *appdb.Handle
	LoginIDConfig      *config.LoginIDConfig
	VerificationConfig *config.VerificationConfig
	Identities         *identityservice.Service
	ControllerFactory  handlerwebapp.ControllerFactory
	BaseViewModel      *viewmodels.BaseViewModeler
	Verification       verification.Service
	Renderer           handlerwebapp.Renderer
	AccountManagement  accountmanagement.Service
}

func (h *AuthflowV2SettingsIdentityViewEmailHandler) GetData(ctx context.Context, r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	loginIDKey := r.Form.Get("q_login_id_key")
	identityID := r.Form.Get("q_identity_id")

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	userID := session.GetUserID(ctx)

	emailIdentity, err := h.Identities.LoginID.Get(ctx, *userID, identityID)
	if err != nil {
		return nil, err
	}

	verified, err := h.AccountManagement.CheckIdentityVerified(ctx, emailIdentity.ToInfo())
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

	vm := AuthflowV2SettingsIdentityViewEmailViewModel{
		LoginIDKey:          loginIDKey,
		EmailIdentity:       emailIdentity,
		Verified:            verified,
		UpdateDisabled:      updateDisabled,
		DeleteDisabled:      deleteDisabled,
		VerificationEnabled: *h.VerificationConfig.Claims.Email.Enabled,
	}
	viewmodels.Embed(data, vm)

	return data, nil
}

func (h *AuthflowV2SettingsIdentityViewEmailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsIdentityViewEmailHTML, data)
		return nil
	})

	ctrl.PostAction("remove", func(ctx context.Context) error {
		err := AuthflowV2SettingsRemoveIdentityEmailSchema.Validator().ValidateValue(ctx, handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		IdentityID := r.Form.Get("x_identity_id")
		_, err = h.AccountManagement.DeleteIdentityEmail(ctx, session.GetSession(ctx), &accountmanagement.DeleteIdentityEmailInput{
			IdentityID: IdentityID,
		})
		if err != nil {
			return err
		}

		loginIDKey := r.Form.Get("q_login_id_key")
		redirectURI, err := url.Parse(AuthflowV2RouteSettingsIdentityListEmail)
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

		err := AuthflowV2SettingsIdentityUpdateVerificationEmailSchema.Validator().ValidateValue(ctx, handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		loginID := r.Form.Get("x_login_id")
		identityID := r.Form.Get("x_identity_id")

		s := session.GetSession(ctx)
		output, err := h.AccountManagement.StartUpdateIdentityEmail(ctx, s, &accountmanagement.StartUpdateIdentityEmailInput{
			LoginID:    loginID,
			LoginIDKey: loginIDKey,
			IdentityID: identityID,
		})
		if err != nil {
			return err
		}

		redirectURI, err := url.Parse(AuthflowV2RouteSettingsIdentityVerifyEmail)

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
