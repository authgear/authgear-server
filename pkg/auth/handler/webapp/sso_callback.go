package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
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
	HandleSSOCallback(w http.ResponseWriter, r *http.Request, providerAlias string) (func(error), error)
}

type SSOCallbackHandler struct {
	Provider  ssoProvider
	TxContext db.TxContext
}

func (h *SSOCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	providerAlias := vars["provider"]

	db.WithTx(h.TxContext, func() error {
		writeResponse, err := h.Provider.HandleSSOCallback(w, r, providerAlias)
		writeResponse(err)
		return err
	})
}
