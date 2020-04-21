package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

func AttachSSOCallbackHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.
		NewRoute().
		Path("/sso/oauth2/callback/{provider}").
		Methods("OPTIONS", "GET", "POST").
		Handler(auth.MakeHandler(authDependency, newSSOCallbackHandler))
}

type ssoProvider interface {
	HandleSSOCallback(w http.ResponseWriter, r *http.Request, oauthProvider webapp.OAuthProvider) (func(error), error)
}

type SSOCallbackHandler struct {
	Provider      ssoProvider
	oauthProvider webapp.OAuthProvider
	TxContext     db.TxContext
}

func (h *SSOCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if h.oauthProvider == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	db.WithTx(h.TxContext, func() error {
		writeResponse, err := h.Provider.HandleSSOCallback(w, r, h.oauthProvider)
		writeResponse(err)
		return err
	})
}
