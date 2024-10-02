package authflowv2

import (
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

var TemplateWebSettingsIdentityViewPhoneHTML = template.RegisterHTML(
	"web/authflowv2/settings_identity_view_phone.html",
	handlerwebapp.SettingsComponents...,
)

var AuthflowV2SettingsUpdateIdentityVerificationPhoneSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_login_id": { "type": "string" },
			"x_identity_id": { "type": "string" }
		},
		"required": ["x_login_id", "x_identity_id"]
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
	LoginIDKey     string
	PhoneIdentity  *identity.LoginID
	Verified       bool
	UpdateDisabled bool
	DeleteDisabled bool
}

type AuthflowV2SettingsIdentityViewPhoneHandler struct {
	Database            *appdb.Handle
	LoginIDConfig       *config.LoginIDConfig
	Identities          *identityservice.Service
	ControllerFactory   handlerwebapp.ControllerFactory
	BaseViewModel       *viewmodels.BaseViewModeler
	Verification        handlerwebapp.SettingsVerificationService
	Renderer            handlerwebapp.Renderer
	AuthenticatorConfig *config.AuthenticatorConfig
	AccountManagement   accountmanagement.Service
}

func (h *AuthflowV2SettingsIdentityViewPhoneHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	loginIDKey := r.Form.Get("q_login_id_key")
	identityID := r.Form.Get("q_identity_id")

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	userID := session.GetUserID(r.Context())

	phoneIdentity, err := h.Identities.LoginID.Get(*userID, identityID)
	if err != nil {
		return nil, err
	}

	verified, err := h.AccountManagement.CheckIdentityVerified(phoneIdentity.ToInfo())
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
		Verified:       verified,
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

	ctrl.PostAction("remove", func() error {
		err := AuthflowV2SettingsRemoveIdentityPhoneSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		removeID := r.Form.Get("x_identity_id")

		_, err = h.AccountManagement.DeleteIdentityPhone(session.GetSession(r.Context()), &accountmanagement.DeleteIdentityPhoneInput{
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

	ctrl.PostAction("verify", func() error {
		loginIDKey := r.Form.Get("q_login_id_key")

		err := AuthflowV2SettingsUpdateIdentityVerificationPhoneSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		loginID := r.Form.Get("x_login_id")
		identityID := r.Form.Get("x_identity_id")

		s := session.GetSession(r.Context())
		output, err := h.AccountManagement.StartUpdateIdentityPhone(s, &accountmanagement.StartUpdateIdentityPhoneInput{
			LoginID:    loginID,
			LoginIDKey: loginIDKey,
			IdentityID: identityID,
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
