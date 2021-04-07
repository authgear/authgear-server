package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/session"
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

type SettingsSessionsViewModel struct {
	CurrentSessionID string
	Sessions         []*model.Session
}

type SettingsSessionsHandler struct {
	ControllerFactory ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	Renderer          Renderer
	Sessions          SettingsSessionManager
}

func (h *SettingsSessionsHandler) GetData(r *http.Request, rw http.ResponseWriter, s session.Session) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	viewModel := SettingsSessionsViewModel{}
	ss, err := h.Sessions.List(s.SessionAttrs().UserID)
	if err != nil {
		return nil, err
	}
	for _, s := range ss {
		viewModel.Sessions = append(viewModel.Sessions, s.ToAPIModel())
	}
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
		ss, err := h.Sessions.List(currentSession.SessionAttrs().UserID)
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
}
