package webapp

import (
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/sessiongroup"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsSessionsHTML = template.RegisterHTML(
	"web/settings_sessions.html",
	components...,
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
	Sessions         []*model.Session
	SessionGroups    []*model.SessionGroup
	Authorizations   []Authorization
}

type SettingsSessionsHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
	Sessions          SettingsSessionManager
	Authorizations    SettingsAuthorizationService
	OAuthConfig       *config.OAuthConfig
}

func (h *SettingsSessionsHandler) GetData(r *http.Request, rw http.ResponseWriter, s session.Session) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	userID := s.GetAuthenticationInfo().UserID
	viewModel := SettingsSessionsViewModel{}
	ss, err := h.Sessions.List(userID)
	if err != nil {
		return nil, err
	}
	for _, s := range ss {
		viewModel.Sessions = append(viewModel.Sessions, s.ToAPIModel())
	}

	// Get third party app authorization
	thirdPartyClientNameMap := map[string]string{}
	for _, c := range h.OAuthConfig.Clients {
		if c.ClientParty() == config.ClientPartyThird {
			thirdPartyClientNameMap[c.ClientID] = c.ClientName
		}
	}
	authorizations, err := h.Authorizations.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	authzs := []Authorization{}
	for _, authz := range authorizations {
		if clientName, ok := thirdPartyClientNameMap[authz.ClientID]; ok {
			authzs = append(authzs, Authorization{
				ID:                    authz.ID,
				ClientID:              authz.ClientID,
				ClientName:            clientName,
				Scope:                 authz.Scopes,
				CreatedAt:             authz.CreatedAt,
				HasFullUserInfoAccess: authz.IsAuthorized([]string{oauth.FullUserInfoScope}),
			})
		}
	}
	viewModel.Authorizations = authzs

	viewModel.CurrentSessionID = s.SessionID()
	viewModel.SessionGroups = sessiongroup.Group(ss)
	viewmodels.Embed(data, viewModel)

	return data, nil
}

func (h *SettingsSessionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

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

		err = h.Sessions.Revoke(s, false)
		if err != nil {
			return err
		}

		result := webapp.Result{RedirectURI: redirectURI}
		result.WriteResponse(w, r)
		return nil
	})

	ctrl.PostAction("revoke_all", func() error {
		ss, err := h.Sessions.List(currentSession.GetAuthenticationInfo().UserID)
		if err != nil {
			return err
		}

		for _, s := range ss {
			if s.SessionID() == currentSession.SessionID() {
				continue
			}
			if err = h.Sessions.Revoke(s, false); err != nil {
				return err
			}
		}

		result := webapp.Result{RedirectURI: redirectURI}
		result.WriteResponse(w, r)
		return nil
	})

	ctrl.PostAction("revoke_group", func() error {
		sessionID := r.Form.Get("x_session_id")

		ss, err := h.Sessions.List(currentSession.GetAuthenticationInfo().UserID)
		if err != nil {
			return err
		}

		// Group the sessions again to find out the target group.
		var targetIDs []string
		groups := sessiongroup.Group(ss)
		for _, group := range groups {
			for _, offlineGrantID := range group.OfflineGrantIDs {
				if sessionID == offlineGrantID {
					for _, s := range group.Sessions {
						if s.ID != currentSession.SessionID() {
							targetIDs = append(targetIDs, s.ID)
						}
					}
				}
			}
		}

		for _, id := range targetIDs {
			for _, s := range ss {
				if s.SessionID() == id {
					err := h.Sessions.Revoke(s, false)
					if err != nil {
						return err
					}
				}
			}
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
