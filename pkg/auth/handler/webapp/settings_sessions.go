package webapp

import (
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/sessionlisting"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsSessionsHTML = template.RegisterHTML(
	"web/settings_sessions.html",
	Components...,
)

func ConfigureSettingsSessionsRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/settings/sessions")
}

type Authorization struct {
	ID                    string
	ClientID              string
	ClientName            string
	Scope                 []string
	CreatedAt             time.Time
	HasFullUserInfoAccess bool
}

type SettingsSessionsViewModel struct {
	CurrentSessionID string
	Sessions         []*sessionlisting.Session
	Authorizations   []Authorization
}

type SettingsSessionsHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
	Sessions          SettingsSessionManager
	Authorizations    SettingsAuthorizationService
	OAuthConfig       *config.OAuthConfig
	SessionListing    SettingsSessionListingService
}

func (h *SettingsSessionsHandler) GetData(r *http.Request, rw http.ResponseWriter, s session.ResolvedSession) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	userID := s.GetAuthenticationInfo().UserID
	viewModel := SettingsSessionsViewModel{}

	ss, err := h.Sessions.List(userID)
	if err != nil {
		return nil, err
	}
	sessionModels, err := h.SessionListing.FilterForDisplay(ss, s)
	if err != nil {
		return nil, err
	}
	viewModel.Sessions = sessionModels

	// Get third party app authorization
	clientNameMap := map[string]string{}
	for _, c := range h.OAuthConfig.Clients {
		clientNameMap[c.ClientID] = c.ClientName
	}
	filter := oauth.NewKeepThirdPartyAuthorizationFilter(h.OAuthConfig)
	authorizations, err := h.Authorizations.ListByUser(userID, filter)
	if err != nil {
		return nil, err
	}
	authzs := []Authorization{}
	for _, authz := range authorizations {
		clientName := clientNameMap[authz.ClientID]
		authzs = append(authzs, Authorization{
			ID:                    authz.ID,
			ClientID:              authz.ClientID,
			ClientName:            clientName,
			Scope:                 authz.Scopes,
			CreatedAt:             authz.CreatedAt,
			HasFullUserInfoAccess: authz.IsAuthorized([]string{oauth.FullUserInfoScope}),
		})
	}
	viewModel.Authorizations = authzs

	viewModel.CurrentSessionID = s.SessionID()
	viewmodels.Embed(data, viewModel)

	return data, nil
}

func (h *SettingsSessionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithDBTx()

	currentSession := session.GetSession(r.Context())
	redirectURI := httputil.HostRelative(r.URL).String()

	ctrl.Get(func() error {
		data, err := h.GetData(r, w, currentSession)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsSessionsHTML, data)
		return nil
	})

	ctrl.PostAction("revoke", func() error {
		sessionID := r.Form.Get("x_session_id")
		if sessionID == currentSession.SessionID() {
			return apierrors.NewInvalid("cannot revoke current session")
		}

		s, err := h.Sessions.Get(sessionID)
		if err != nil {
			return err
		}

		err = h.Sessions.RevokeWithEvent(s, true, false)
		if err != nil {
			return err
		}

		result := webapp.Result{RedirectURI: redirectURI}
		result.WriteResponse(w, r)
		return nil
	})

	ctrl.PostAction("revoke_all", func() error {
		userID := currentSession.GetAuthenticationInfo().UserID
		err := h.Sessions.TerminateAllExcept(userID, currentSession, false)
		if err != nil {
			return err
		}

		result := webapp.Result{RedirectURI: redirectURI}
		result.WriteResponse(w, r)
		return nil
	})

	ctrl.PostAction("remove_authorization", func() error {
		authorizationID := r.Form.Get("x_authorization_id")
		authz, err := h.Authorizations.GetByID(authorizationID)
		if err != nil {
			return err
		}

		if authz.UserID != currentSession.GetAuthenticationInfo().UserID {
			return apierrors.NewForbidden("cannot remove authorization")
		}

		err = h.Authorizations.Delete(authz)
		if err != nil {
			return err
		}

		result := webapp.Result{RedirectURI: redirectURI}
		result.WriteResponse(w, r)
		return nil
	})
}
