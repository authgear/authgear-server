package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebLogoutHTML = template.RegisterHTML(
	"web/logout.html",
	Components...,
)

func ConfigureLogoutRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/logout")
}

type LogoutSessionManager interface {
	Logout(session.ListableSession, http.ResponseWriter) error
}

type LogoutHandler struct {
	ControllerFactory   ControllerFactory
	Database            *appdb.Handle
	TrustProxy          config.TrustProxy
	OAuth               *config.OAuthConfig
	UIConfig            *config.UIConfig
	SessionManager      LogoutSessionManager
	BaseViewModel       *viewmodels.BaseViewModeler
	Renderer            Renderer
	OAuthClientResolver WebappOAuthClientResolver
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	ctrl.Get(func() error {
		baseViewModel := h.BaseViewModel.ViewModel(r, w)

		data := map[string]interface{}{}

		viewmodels.Embed(data, baseViewModel)

		h.Renderer.RenderHTML(w, r, TemplateWebLogoutHTML, data)
		return nil
	})

	ctrl.PostAction("logout", func() error {
		sess := session.GetSession(r.Context())
		err := h.SessionManager.Logout(sess, w)
		if err != nil {
			return err
		}

		uiParam := uiparam.GetUIParam(r.Context())
		clientID := uiParam.ClientID
		client := h.OAuthClientResolver.ResolveClient(clientID)
		postLogoutRedirectURI := webapp.ResolvePostLogoutRedirectURI(client, r.FormValue("post_logout_redirect_uri"), h.UIConfig)
		redirectURI := webapp.GetRedirectURI(r, bool(h.TrustProxy), postLogoutRedirectURI)
		http.Redirect(w, r, redirectURI, http.StatusFound)
		return nil
	})
}
