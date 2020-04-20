package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
)

func AttachForgotPasswordHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.
		NewRoute().
		Path("/forgot_password").
		Handler(auth.MakeHandler(authDependency, newForgotPasswordHandler))
}

type ForgotPasswordHandler struct {
	RenderProvider webapp.RenderProvider
}

func (h *ForgotPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.RenderProvider.WritePage(w, r, webapp.TemplateItemTypeAuthUIForgotPasswordHTML, map[string]interface{}{})
}
