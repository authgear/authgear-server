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
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebSettingsIdentityEditEmailHTML = template.RegisterHTML(
	"web/authflowv2/settings_identity_edit_email.html",
	handlerwebapp.SettingsComponents...,
)

var AuthflowV2SettingsIdentityEditEmailSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_login_id": { "type": "string" },
			"x_identity_id": { "type": "string" }
		},
		"required": ["x_login_id", "x_identity_id"]
	}
`)

func ConfigureAuthflowV2SettingsIdentityEditEmailRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsIdentityChangeEmail)
}

type AuthflowV2SettingsIdentityEditEmailViewModel struct {
	LoginIDKey string
	IdentityID string
	Target     *identity.LoginID
}

type AuthflowV2SettingsIdentityEditEmailHandler struct {
	Database          *appdb.Handle
	ControllerFactory handlerwebapp.ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          handlerwebapp.Renderer
	AccountManagement accountmanagement.Service
	Identities        *identityservice.Service
}

func (h *AuthflowV2SettingsIdentityEditEmailHandler) GetData(ctx context.Context, r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	loginIDKey := r.Form.Get("q_login_id_key")
	identityID := r.Form.Get("q_identity_id")

	userID := session.GetUserID(ctx)

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	target, err := h.Identities.LoginID.Get(ctx, *userID, identityID)
	if err != nil {
		return nil, err
	}

	vm := AuthflowV2SettingsIdentityEditEmailViewModel{
		LoginIDKey: loginIDKey,
		IdentityID: identityID,
		Target:     target,
	}
	viewmodels.Embed(data, vm)

	return data, nil
}

func (h *AuthflowV2SettingsIdentityEditEmailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx(r.Context())

	ctrl.Get(func(ctx context.Context) error {
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

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsIdentityEditEmailHTML, data)
		return nil
	})

	ctrl.PostAction("", func(ctx context.Context) error {

		loginIDKey := r.Form.Get("q_login_id_key")

		err := AuthflowV2SettingsIdentityEditEmailSchema.Validator().ValidateValue(handlerwebapp.FormToJSON(r.Form))
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

		var redirectURI *url.URL
		if output.NeedVerification {
			redirectURI, err = url.Parse(AuthflowV2RouteSettingsIdentityVerifyEmail)

			q := redirectURI.Query()
			q.Set("q_login_id_key", loginIDKey)
			q.Set("q_token", output.Token)

			redirectURI.RawQuery = q.Encode()
		} else {
			redirectURI, err = url.Parse(AuthflowV2RouteSettingsIdentityListEmail)

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
