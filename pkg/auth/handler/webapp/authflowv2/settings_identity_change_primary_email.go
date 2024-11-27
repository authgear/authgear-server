package authflowv2

import (
	"context"
	"net/http"

	"net/url"

	"github.com/authgear/authgear-server/pkg/api/model"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/setutil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsIdentityChangePrimaryEmailHTML = template.RegisterHTML(
	"web/authflowv2/settings_identity_change_primary_email.html",
	handlerwebapp.SettingsComponents...,
)

func ConfigureAuthflowV2SettingsIdentityChangePrimaryEmailRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsIdentityChangePrimaryEmail)
}

type AuthflowV2SettingsIdentityChangePrimaryEmailViewModel struct {
	LoginIDKey string
	Emails     []string
}

type AuthflowV2SettingsIdentityChangePrimaryEmailHandler struct {
	Database                 *appdb.Handle
	ControllerFactory        handlerwebapp.ControllerFactory
	BaseViewModel            *viewmodels.BaseViewModeler
	SettingsProfileViewModel *viewmodels.SettingsProfileViewModeler
	Renderer                 handlerwebapp.Renderer
	Users                    handlerwebapp.SettingsProfileEditUserService
	StdAttrs                 handlerwebapp.SettingsProfileEditStdAttrsService
	Identities               *identityservice.Service
}

func (h *AuthflowV2SettingsIdentityChangePrimaryEmailHandler) GetData(ctx context.Context, r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	loginIDKey := r.Form.Get("q_login_id_key")

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	userID := session.GetUserID(ctx)

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
	viewmodels.Embed(data, *settingsProfileViewModel)

	emails := setutil.Set[string]{}
	for _, identity := range loginIDIdentities {
		if identity.LoginIDType == model.LoginIDKeyTypeEmail {
			emails.Add(identity.LoginID)
		}
	}

	for _, identity := range oauthIdentities {
		email, ok := identity.Claims[stdattrs.Email].(string)
		if ok && email != "" {
			emails.Add(email)
		}
	}

	vm := AuthflowV2SettingsIdentityChangePrimaryEmailViewModel{
		LoginIDKey: loginIDKey,
		Emails:     emails.Keys(),
	}
	viewmodels.Embed(data, vm)

	return data, nil
}

func (h *AuthflowV2SettingsIdentityChangePrimaryEmailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsIdentityChangePrimaryEmailHTML, data)
		return nil
	})

	ctrl.PostAction("save", func(ctx context.Context) error {
		userID := *session.GetUserID(ctx)
		m := handlerwebapp.JSONPointerFormToMap(r.Form)

		err := h.Database.WithTx(ctx, func(ctx context.Context) error {
			u, err := h.Users.GetRaw(ctx, userID)
			if err != nil {
				return err
			}

			attrs, err := stdattrs.T(u.StandardAttributes).MergedWithForm(m)
			if err != nil {
				return err
			}

			err = h.StdAttrs.UpdateStandardAttributes(ctx, config.RoleEndUser, userID, attrs)
			if err != nil {
				return err
			}

			return nil
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
}
