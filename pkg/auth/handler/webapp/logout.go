package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebLogoutHTML = template.RegisterHTML(
	"web/logout.html",
	components...,
)

func ConfigureLogoutRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/logout")
}

type LogoutSessionManager interface {
	Logout(session.Session, http.ResponseWriter) error
}

type LogoutHandler struct {
	Database       *db.Handle
	TrustProxy     config.TrustProxy
	OAuth          *config.OAuthConfig
	UIConfig       *config.UIConfig
	SessionManager LogoutSessionManager
	BaseViewModel  *viewmodels.BaseViewModeler
	Renderer       Renderer
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		baseViewModel := h.BaseViewModel.ViewModel(r, w)

		data := map[string]interface{}{}

		viewmodels.Embed(data, baseViewModel)

		h.Renderer.RenderHTML(w, r, TemplateWebLogoutHTML, data)
		return
	}

	if r.Method == "POST" {
		err := h.Database.WithTx(func() error {
			sess := session.GetSession(r.Context())
			err := h.SessionManager.Logout(sess, w)
			if err != nil {
				return err
			}

			clientID := webapp.GetClientID(r.Context())
			client, _ := h.OAuth.GetClient(clientID)
			postLogoutRedirectURI := webapp.ResolvePostLogoutRedirectURI(client, r.FormValue("post_logout_redirect_uri"), h.UIConfig)
			redirectURI := webapp.GetRedirectURI(r, bool(h.TrustProxy), postLogoutRedirectURI)
			http.Redirect(w, r, redirectURI, http.StatusFound)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
}
