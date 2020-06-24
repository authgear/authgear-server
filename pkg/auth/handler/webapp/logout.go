package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
	"github.com/skygeario/skygear-server/pkg/db"
)

func ConfigureLogoutHandler(router *mux.Router, h http.Handler) {
	router.NewRoute().
		Path("/logout").
		Methods("OPTIONS", "POST", "GET").
		Handler(h)
}

type LogoutSessionManager interface {
	Logout(auth.AuthSession, http.ResponseWriter) error
}

type LogoutHandler struct {
	ServerConfig   *config.ServerConfig
	RenderProvider webapp.RenderProvider
	SessionManager LogoutSessionManager
	DBContext      db.Context
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	db.WithTx(h.DBContext, func() error {
		if r.Method == "POST" && r.Form.Get("x_action") == "logout" {
			sess := auth.GetSession(r.Context())
			h.SessionManager.Logout(sess, w)
			webapp.RedirectToRedirectURI(w, r, h.ServerConfig.TrustProxy)
		} else {
			h.RenderProvider.WritePage(w, r, webapp.TemplateItemTypeAuthUILogoutHTML, nil)
		}
		return nil
	})
}
