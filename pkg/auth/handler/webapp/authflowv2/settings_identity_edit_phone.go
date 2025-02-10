package authflowv2

import (
	"context"
	"errors"
	"net/http"

	"net/url"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/stringutil"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebSettingsIdentityEditPhoneHTML = template.RegisterHTML(
	"web/authflowv2/settings_identity_edit_phone.html",
	handlerwebapp.SettingsComponents...,
)

var AuthflowV2SettingsIdentityEditPhoneSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_login_id_key": { "type": "string" },
			"x_channel": { "type": "string" },
			"x_login_id": { "type": "string" },
			"x_identity_id": { "type": "string" }
		},
		"required": ["x_login_id_key", "x_channel", "x_login_id", "x_identity_id"]
	}
`)

func ConfigureAuthflowV2SettingsIdentityEditPhoneRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsIdentityChangePhone)
}

type AuthflowV2SettingsIdentityEditPhoneViewModel struct {
	LoginIDKey string
	Channel    model.AuthenticatorOOBChannel
	IdentityID string
	Target     *identity.LoginID
}

type AuthflowV2SettingsIdentityEditPhoneHandler struct {
	Database            *appdb.Handle
	ControllerFactory   handlerwebapp.ControllerFactory
	BaseViewModel       *viewmodels.BaseViewModeler
	Renderer            handlerwebapp.Renderer
	AuthenticatorConfig *config.AuthenticatorConfig
	AccountManagement   accountmanagement.Service
	Identities          *identityservice.Service
}

func (h *AuthflowV2SettingsIdentityEditPhoneHandler) GetData(ctx context.Context, r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	loginIDKey := r.Form.Get("q_login_id_key")
	identityID := r.Form.Get("q_identity_id")
	loginIDValue := r.Form.Get("q_login_id")

	userID := session.GetUserID(ctx)

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	channel := h.AuthenticatorConfig.OOB.SMS.PhoneOTPMode.GetDefaultChannel()

	var target *identity.Info
	var err error

	if identityID != "" {
		target, err = h.Identities.GetWithUserID(ctx, *userID, identityID)
		if err != nil {
			return nil, err
		}
	} else if loginIDValue != "" {
		target, err = h.Identities.GetBySpecWithUserID(ctx, *userID, &identity.Spec{
			Type: model.IdentityTypeLoginID,
			LoginID: &identity.LoginIDSpec{
				Key:   loginIDKey,
				Type:  model.LoginIDKeyTypePhone,
				Value: stringutil.NewUserInputString(loginIDValue),
			},
		})
		if err != nil {
			return nil, err
		}
	} else {
		// No query parameter provided, treat as not found
		err = api.ErrIdentityNotFound
	}

	if err != nil && errors.Is(err, api.ErrIdentityNotFound) {
		return nil, apierrors.AddDetails(err, errorutil.Details{
			"LoginIDType": model.LoginIDKeyTypePhone,
		})
	} else if err != nil {
		return nil, err
	}

	vm := AuthflowV2SettingsIdentityEditPhoneViewModel{
		LoginIDKey: loginIDKey,
		Channel:    channel,
		IdentityID: identityID,
		Target:     target.LoginID,
	}
	viewmodels.Embed(data, vm)

	return data, nil
}

func (h *AuthflowV2SettingsIdentityEditPhoneHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx(r.Context())

	ctrl.GetWithSettingsActionWebSession(r, func(ctx context.Context, _ *webapp.Session) error {
		var data map[string]interface{}
		err = h.Database.WithTx(ctx, func(ctx context.Context) error {
			data, err = h.GetData(ctx, r, w)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsIdentityEditPhoneHTML, data)
		return nil
	})

	ctrl.PostActionWithSettingsActionWebSession("", r, func(ctx context.Context, webappSession *webapp.Session) error {

		err := AuthflowV2SettingsIdentityEditPhoneSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}

		channel := model.AuthenticatorOOBChannel(r.Form.Get("x_channel"))
		loginIDKey := r.Form.Get("x_login_id_key")
		loginID := r.Form.Get("x_login_id")
		identityID := r.Form.Get("x_identity_id")

		s := session.GetSession(ctx)

		output, err := h.AccountManagement.StartUpdateIdentityPhone(ctx, s, &accountmanagement.StartUpdateIdentityPhoneInput{
			Channel:    channel,
			LoginID:    loginID,
			LoginIDKey: loginIDKey,
			IdentityID: identityID,
		})
		if err != nil {
			return err
		}

		var redirectURI *url.URL
		if output.NeedVerification {
			redirectURI, err = url.Parse(AuthflowV2RouteSettingsIdentityVerifyPhone)

			q := redirectURI.Query()
			q.Set("q_login_id_key", loginIDKey)
			q.Set("q_token", output.Token)

			redirectURI.RawQuery = q.Encode()
		} else {
			if ctrl.IsInSettingsAction(s, webappSession) {
				settingsActionResult, err := ctrl.FinishSettingsActionWithResult(ctx, s, webappSession)
				if err != nil {
					return err
				}
				settingsActionResult.WriteResponse(w, r)
				return nil
			}

			redirectURI, err = url.Parse(AuthflowV2RouteSettingsIdentityListPhone)

			q := redirectURI.Query()
			q.Set("q_login_id_key", loginIDKey)

			redirectURI.RawQuery = q.Encode()
		}
		if err != nil {
			return err
		}

		result := webapp.Result{RedirectURI: redirectURI.String()}
		result.WriteResponse(w, r)
		return nil
	})
}
