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
	SessionManager LogoutSessionManager
	BaseViewModel  *viewmodels.BaseViewModeler
	Renderer       Renderer
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		baseViewModel := h.BaseViewModel.ViewModel(r, nil)

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

			redirectURI := webapp.GetRedirectURI(r, bool(h.TrustProxy), "/settings")
			http.Redirect(w, r, redirectURI, http.StatusFound)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
}
