package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/db"
	"github.com/skygeario/skygear-server/pkg/deps"
)

func AttachSSOCallbackHandler(
	router *mux.Router,
	p *deps.RootProvider,
) {
	router.
		NewRoute().
		Path("/sso/oauth2/callback/{provider}").
		Methods("OPTIONS", "GET", "POST").
		Handler(p.Handler(newSSOCallbackHandler))
}

type ssoProvider interface {
	HandleSSOCallback(w http.ResponseWriter, r *http.Request, providerAlias string) (func(error), error)
}

type SSOCallbackHandler struct {
	Provider  ssoProvider
	DBContext db.Context
}

func (h *SSOCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	providerAlias := vars["provider"]

	db.WithTx(h.DBContext, func() error {
		writeResponse, err := h.Provider.HandleSSOCallback(w, r, providerAlias)
		writeResponse(err)
		return err
	})
}
