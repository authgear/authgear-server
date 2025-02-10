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
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/stringutil"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateSettingsIdentityEditUsernameTemplate = template.RegisterHTML(
	"web/authflowv2/settings_identity_edit_username.html",
	handlerwebapp.SettingsComponents...,
)

var AuthflowV2SettingsIdentityEditUsernameSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_login_id_key": { "type": "string" },
			"x_login_id": { "type": "string" },
			"x_identity_id": { "type": "string" }
		},
		"required": ["x_login_id_key", "x_login_id", "x_identity_id"]
	}
`)

func ConfigureAuthflowV2SettingsIdentityEditUsername(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsIdentityChangeUsername)
}

type AuthflowV2SettingsIdentityEditUsernameViewModel struct {
	Identity *identity.LoginID
}

type AuthflowV2SettingsIdentityEditUsernameHandler struct {
	Database          *appdb.Handle
	AccountManagement *accountmanagement.Service
	Identities        *identityservice.Service
	ControllerFactory handlerwebapp.ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          handlerwebapp.Renderer
}

func (h *AuthflowV2SettingsIdentityEditUsernameHandler) GetData(ctx context.Context, w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, w)
	viewmodels.Embed(data, baseViewModel)

	userID := session.GetUserID(ctx)
	loginIDKey := r.Form.Get("q_login_id_key")
	identityID := r.Form.Get("q_identity_id")
	loginIDValue := r.Form.Get("q_login_id")

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
				Type:  model.LoginIDKeyTypeUsername,
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
			"LoginIDType": model.LoginIDKeyTypeUsername,
		})
	} else if err != nil {
		return nil, err
	}

	vm := AuthflowV2SettingsIdentityEditUsernameViewModel{
		Identity: target.LoginID,
	}
	viewmodels.Embed(data, vm)

	return data, nil
}

func (h *AuthflowV2SettingsIdentityEditUsernameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx(r.Context())

	ctrl.GetWithSettingsActionWebSession(r, func(ctx context.Context, _ *webapp.Session) error {
		var data map[string]interface{}
		err := h.Database.WithTx(ctx, func(ctx context.Context) error {
			data, err = h.GetData(ctx, w, r)
			return err
		})
		if err != nil {
			return err
		}
		h.Renderer.RenderHTML(w, r, TemplateSettingsIdentityEditUsernameTemplate, data)
		return nil
	})

	ctrl.PostActionWithSettingsActionWebSession("", r, func(ctx context.Context, webappSession *webapp.Session) error {

		s := session.GetSession(ctx)
		err = AuthflowV2SettingsIdentityEditUsernameSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
		if err != nil {
			return err
		}
		identityID := r.Form.Get("x_identity_id")
		loginIDKey := r.Form.Get("x_login_id_key")
		loginID := r.Form.Get("x_login_id")
		resolvedSession := session.GetSession(ctx)
		_, err = h.AccountManagement.UpdateIdentityUsername(ctx, resolvedSession, &accountmanagement.UpdateIdentityUsernameInput{
			IdentityID: identityID,
			LoginIDKey: loginIDKey,
			LoginID:    loginID,
		})
		if err != nil {
			return err
		}

		if ctrl.IsInSettingsAction(s, webappSession) {
			settingsActionResult, err := ctrl.FinishSettingsActionWithResult(ctx, s, webappSession)
			if err != nil {
				return err
			}
			settingsActionResult.WriteResponse(w, r)
			return nil
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
