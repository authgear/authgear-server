package authflowv2

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/resourcescope"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsV2AuthorizedAppsHTML = template.RegisterHTML(
	"web/authflowv2/settings_authorized_apps.html",
	handlerwebapp.SettingsComponents...,
)

var TemplateWebSettingsV2AuthorizedAppHTML = template.RegisterHTML(
	"web/authflowv2/settings_authorized_app.html",
	handlerwebapp.SettingsComponents...,
)

const authorizedAppsListPath = "/settings/authorized-apps"

func ConfigureAuthflowV2SettingsAuthorizedAppsRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern(authorizedAppsListPath)
}

// The detail page of one authorization: what the app can access, and the
// action to remove it. Reached from the list by q_authorization_id.
func ConfigureAuthflowV2SettingsAuthorizedAppRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/authorized-apps/view")
}

type SettingsSessionsClientResolver interface {
	ResolveClient(ctx context.Context, clientID string) *config.OAuthClientConfig
}

// SettingsAuthorizedAppsClientLogoEndpoint is a narrow interface (not the
// whole oauth.EndpointsProvider) so this handler's dependency is exactly the
// one method it uses -- same shape as oauth/consent.go's
// ConsentClientLogoEndpoint.
type SettingsAuthorizedAppsClientLogoEndpoint interface {
	ClientLogoURL(clientID string) *url.URL
}

// SettingsAuthorizedAppsScopeStore resolves the scope names stored on an
// authorization back to the project's configured scopes, for their
// descriptions.
type SettingsAuthorizedAppsScopeStore interface {
	ListScopesByNames(ctx context.Context, names []string) ([]*resourcescope.Scope, error)
}

// AuthorizationPermission is one granted resource scope as shown to the end
// user. DisplayText is the scope's configured Description, or the raw scope
// name when none is configured or the scope no longer exists, so a granted
// permission is never silently omitted -- the same rule the consent screen
// applies (see oauth/consent.go's ConsentScope).
type AuthorizationPermission struct {
	Scope       string
	DisplayText string
}

type Authorization struct {
	ID            string
	ClientID      string
	ClientName    string
	ClientLogoURI string
	Scope         []string
	// Permissions are the granted resource scopes (oauth.IsResourceScope), in
	// the order they were granted; set on the detail page only. Project-level
	// scopes are not in this list: the template renders the identity ones
	// (profile, email, phone, address, full-userinfo) from its own strings.
	Permissions []AuthorizationPermission
	CreatedAt   time.Time
}

type SettingsAuthorizedAppsViewModel struct {
	Authorizations []Authorization
}

type SettingsAuthorizedAppViewModel struct {
	Authorization Authorization
}

type AuthflowV2SettingsAuthorizedAppsHandler struct {
	Database            *appdb.Handle
	ControllerFactory   handlerwebapp.ControllerFactory
	BaseViewModel       *viewmodels.BaseViewModeler
	SettingsViewModel   *viewmodels.SettingsViewModeler
	Renderer            handlerwebapp.Renderer
	Authorizations      SettingsAuthorizationService
	OAuthClientResolver SettingsSessionsClientResolver
	Endpoints           SettingsAuthorizedAppsClientLogoEndpoint
	ScopeStore          SettingsAuthorizedAppsScopeStore
}

// resourcePermissions maps the resource scopes on granted (in grant order)
// to their display text. A name defined by more than one resource keeps the
// first description found: the grant does not record which resource it was
// for.
func resourcePermissions(granted []string, descriptions map[string]string) []AuthorizationPermission {
	var permissions []AuthorizationPermission
	for _, s := range granted {
		if !oauth.IsResourceScope(s) {
			continue
		}
		displayText := s
		if d, ok := descriptions[s]; ok && d != "" {
			displayText = d
		}
		permissions = append(permissions, AuthorizationPermission{Scope: s, DisplayText: displayText})
	}
	return permissions
}

// scopeDescriptions returns Description keyed by scope name for the resource
// scopes in granted. A scope that no longer exists has no entry, and is shown
// by its raw name.
func (h *AuthflowV2SettingsAuthorizedAppsHandler) scopeDescriptions(ctx context.Context, granted []string) (map[string]string, error) {
	var names []string
	for _, s := range granted {
		if oauth.IsResourceScope(s) {
			names = append(names, s)
		}
	}
	if len(names) == 0 {
		return nil, nil
	}
	scopes, err := h.ScopeStore.ListScopesByNames(ctx, names)
	if err != nil {
		return nil, err
	}
	descriptions := make(map[string]string, len(scopes))
	for _, sc := range scopes {
		if sc.Description == nil || *sc.Description == "" {
			continue
		}
		if _, exists := descriptions[sc.Scope]; exists {
			continue
		}
		descriptions[sc.Scope] = *sc.Description
	}
	return descriptions, nil
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
		authzs = append(authzs, h.authorizationViewModel(ctx, authz))
	}

	settingsAuthorizedAppsViewModel := SettingsAuthorizedAppsViewModel{
		Authorizations: authzs,
	}
	viewmodels.Embed(data, settingsAuthorizedAppsViewModel)

	return data, nil
}

// authorizationViewModel resolves the client behind an authorization for
// display: its name and its logo (proxied for a dynamic client). Permissions
// are left to the detail page, the only one that shows them.
func (h *AuthflowV2SettingsAuthorizedAppsHandler) authorizationViewModel(ctx context.Context, authz *oauth.Authorization) Authorization {
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
		// Point the <img> at Authgear's own proxy instead of the
		// client's server, so the end user's browser never contacts
		// the client (spec § Privacy Considerations §9.2). Only for a
		// dynamic client, matching oauth/consent.go's
		// consentViewModelForClient -- a static client's logo is
		// project-collaborator-configured and continues to render
		// directly.
		if c.LogoURI != "" && c.IsDynamicClient() {
			logoURI = h.Endpoints.ClientLogoURL(c.ClientID).String()
		} else {
			logoURI = c.LogoURI
		}
	}
	return Authorization{
		ID:            authz.ID,
		ClientID:      authz.ClientID,
		ClientName:    clientName,
		ClientLogoURI: logoURI,
		Scope:         authz.Scopes,
		CreatedAt:     authz.CreatedAt,
	}
}

// getAuthorizationForUser loads the authorization named by q_authorization_id
// and checks it belongs to the signed-in user. Any other user's grant -- or a
// first-party one, which this page does not list -- is reported as not found
// rather than forbidden, so the page does not confirm that the id exists.
func (h *AuthflowV2SettingsAuthorizedAppsHandler) getAuthorizationForUser(ctx context.Context, userID string, authorizationID string) (*oauth.Authorization, error) {
	if authorizationID == "" {
		return nil, apierrors.NewNotFound("authorization not found")
	}
	authz, err := h.Authorizations.GetByID(ctx, authorizationID)
	if errors.Is(err, oauth.ErrAuthorizationNotFound) {
		return nil, apierrors.NewNotFound("authorization not found")
	}
	if err != nil {
		return nil, err
	}
	if authz.UserID != userID {
		return nil, apierrors.NewNotFound("authorization not found")
	}
	// Same rule as the list's KeepThirdPartyAuthorizationFilter: a grant
	// whose client no longer resolves, or which is first-party, is not part
	// of this surface.
	if !oauth.NewKeepThirdPartyAuthorizationFilter(h.OAuthClientResolver).Keep(ctx, authz) {
		return nil, apierrors.NewNotFound("authorization not found")
	}
	return authz, nil
}

// GetAuthorizationData is the detail page's view model: one authorization.
func (h *AuthflowV2SettingsAuthorizedAppsHandler) GetAuthorizationData(ctx context.Context, r *http.Request, rw http.ResponseWriter, authorizationID string) (map[string]any, error) {
	data := map[string]any{}
	userID := session.GetUserID(ctx)

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	settingsViewModel, err := h.SettingsViewModel.ViewModel(ctx, *userID)
	if err != nil {
		return nil, err
	}
	viewmodels.Embed(data, *settingsViewModel)

	authz, err := h.getAuthorizationForUser(ctx, *userID, authorizationID)
	if err != nil {
		return nil, err
	}
	descriptions, err := h.scopeDescriptions(ctx, authz.Scopes)
	if err != nil {
		return nil, err
	}

	vm := h.authorizationViewModel(ctx, authz)
	vm.Permissions = resourcePermissions(authz.Scopes, descriptions)
	viewmodels.Embed(data, SettingsAuthorizedAppViewModel{Authorization: vm})

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
}

// AuthflowV2SettingsAuthorizedAppHandler serves one authorization's detail
// page and its remove action. It shares the list handler's dependencies and
// resolution helpers.
type AuthflowV2SettingsAuthorizedAppHandler struct {
	AuthflowV2SettingsAuthorizedAppsHandler
}

func (h *AuthflowV2SettingsAuthorizedAppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx(r.Context())

	currentSession := session.GetSession(r.Context())
	authorizationID := r.Form.Get("q_authorization_id")

	ctrl.Get(func(ctx context.Context) error {
		var data map[string]any
		err := h.Database.WithTx(ctx, func(ctx context.Context) error {
			data, err = h.GetAuthorizationData(ctx, r, w, authorizationID)
			return err
		})
		if apierrors.IsAPIErrorWithCondition(err, func(e *apierrors.APIError) bool {
			return e.Name == apierrors.NotFound
		}) {
			// The usual way to reach a missing grant is the back button after
			// removing the app. Return to the list rather than showing an
			// error page for what the user just did on purpose.
			http.Redirect(w, r, authorizedAppsListPath, http.StatusFound)
			return nil
		}
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsV2AuthorizedAppHTML, data)

		return nil
	})

	ctrl.PostAction("remove_authorization", func(ctx context.Context) error {
		err := h.Database.WithTx(ctx, func(ctx context.Context) error {
			authz, err := h.getAuthorizationForUser(ctx, currentSession.GetAuthenticationInfo().UserID, authorizationID)
			if err != nil {
				return err
			}
			return h.Authorizations.Delete(ctx, authz)
		})
		if err != nil {
			return err
		}

		// Back to the list: the page just removed is gone.
		result := webapp.Result{RedirectURI: authorizedAppsListPath}
		result.WriteResponse(w, r)
		return nil
	})
}
