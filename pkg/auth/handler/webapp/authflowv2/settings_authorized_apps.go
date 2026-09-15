package authflowv2

import (
	"context"
	"net/http"

	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsV2AuthorizedAppsHTML = template.RegisterHTML(
	"web/authflowv2/settings_authorized_apps.html",
	handlerwebapp.SettingsComponents...,
)

func ConfigureAuthflowV2SettingsAuthorizedAppsRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/authorized-apps")
}

type SettingsSessionsClientResolver interface {
	ResolveClient(ctx context.Context, clientID string) *config.OAuthClientConfig
}

type Authorization struct {
	ID            string
	ClientID      string
	ClientName    string
	ClientLogoURI string
	Scope         []string
	CreatedAt     time.Time
}

type SettingsAuthorizedAppsViewModel struct {
	Authorizations []Authorization
}

type AuthflowV2SettingsAuthorizedAppsHandler struct {
	Database            *appdb.Handle
	ControllerFactory   handlerwebapp.ControllerFactory
	BaseViewModel       *viewmodels.BaseViewModeler
	SettingsViewModel   *viewmodels.SettingsViewModeler
	Renderer            handlerwebapp.Renderer
	Authorizations      SettingsAuthorizationService
	OAuthClientResolver SettingsSessionsClientResolver
}

func (h *AuthflowV2SettingsAuthorizedAppsHandler) GetData(ctx context.Context, r *http.Request, rw http.ResponseWriter, s session.ResolvedSession) (map[string]any, error) {
	data := map[string]any{}
	userID := session.GetUserID(ctx)

	// BaseViewModel
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	// SettingsViewModel
	settingsViewModel, err := h.SettingsViewModel.ViewModel(ctx, *userID)
	if err != nil {
		return nil, err
	}
	viewmodels.Embed(data, *settingsViewModel)

	// Get third party app authorization
	filter := oauth.NewKeepThirdPartyAuthorizationFilter(h.OAuthClientResolver)
	authorizations, err := h.Authorizations.ListByUser(ctx, *userID, filter)
	if err != nil {
		return nil, err
	}
	authzs := []Authorization{}
	for _, authz := range authorizations {
		// One resolve per authorization, same as the filter just did --
		// ResolveClient is cached, and the alternative (threading the
		// resolved config out of the filter) would couple the filter to
		// this caller's rendering needs.
		//
		// Client.Name, not Client.ClientName: the display-name fallback, so
		// a dynamic client with no client_name shows "Client <clientID>"
		// rather than blank. Same fix as consent.go's consentViewModelForClient.
		clientName := authz.ClientID
		logoURI := ""
		if c := h.OAuthClientResolver.ResolveClient(ctx, authz.ClientID); c != nil {
			clientName = c.Name
			logoURI = c.LogoURI
		}
		authzs = append(authzs, Authorization{
			ID:            authz.ID,
			ClientID:      authz.ClientID,
			ClientName:    clientName,
			ClientLogoURI: logoURI,
			Scope:         authz.Scopes,
			CreatedAt:     authz.CreatedAt,
		})
	}

	settingsAuthorizedAppsViewModel := SettingsAuthorizedAppsViewModel{
		Authorizations: authzs,
	}
	viewmodels.Embed(data, settingsAuthorizedAppsViewModel)

	return data, nil
}

func (h *AuthflowV2SettingsAuthorizedAppsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx(r.Context())

	currentSession := session.GetSession(r.Context())
	redirectURI := httputil.HostRelative(r.URL).String()

	ctrl.Get(func(ctx context.Context) error {
		var data map[string]any
		err := h.Database.WithTx(ctx, func(ctx context.Context) error {
			data, err = h.GetData(ctx, r, w, currentSession)
			return err
		})
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsV2AuthorizedAppsHTML, data)

		return nil
	})

	ctrl.PostAction("remove_authorization", func(ctx context.Context) error {
		authorizationID := r.Form.Get("x_authorization_id")
		err := h.Database.WithTx(ctx, func(ctx context.Context) error {
			authz, err := h.Authorizations.GetByID(ctx, authorizationID)
			if err != nil {
				return err
			}

			if authz.UserID != currentSession.GetAuthenticationInfo().UserID {
				return apierrors.NewForbidden("cannot remove authorization")
			}

			err = h.Authorizations.Delete(ctx, authz)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return err
		}

		result := webapp.Result{RedirectURI: redirectURI}
		result.WriteResponse(w, r)
		return nil
	})
}
