package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
)

func AttachLogoutHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.
		NewRoute().
		Path("/logout").
		Handler(pkg.MakeHandler(authDependency, newLogoutHandler))
}

type logoutSessionManager interface {
	Logout(auth.AuthSession, http.ResponseWriter) error
}

type LogoutHandler struct {
	RenderProvider webapp.RenderProvider
	SessionManager logoutSessionManager
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" && r.Form.Get("x_action") == "logout" {
		sess := auth.GetSession(r.Context())
		h.SessionManager.Logout(sess, w)
		webapp.RedirectToRedirectURI(w, r)
	} else {
		h.RenderProvider.WritePage(w, r, webapp.TemplateItemTypeAuthUILogoutHTML, nil)
	}
}
