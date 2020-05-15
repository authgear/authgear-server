package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

func AttachPromoteHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.
		NewRoute().
		Path("/promote_user").
		Methods("OPTIONS", "POST", "GET").
		Handler(auth.MakeHandler(authDependency, newPromoteHandler))
}

type promoteProvider interface {
	GetPromoteLoginIDForm(w http.ResponseWriter, r *http.Request) (func(error), error)
	PromoteLoginID(w http.ResponseWriter, r *http.Request) (func(error), error)
	PromoteIdentityProvider(w http.ResponseWriter, r *http.Request, providerAlias string) (func(error), error)
}

type PromoteHandler struct {
	Provider  promoteProvider
	TxContext db.TxContext
}

func (h *PromoteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db.WithTx(h.TxContext, func() error {
		if r.Method == "GET" {
			writeResponse, err := h.Provider.GetPromoteLoginIDForm(w, r)
			writeResponse(err)
			return err
		}

		if r.Method == "POST" {
			if r.Form.Get("x_idp_id") != "" {
				writeResponse, err := h.Provider.PromoteIdentityProvider(w, r, r.Form.Get("x_idp_id"))
				writeResponse(err)
				return err
			}

			writeResponse, err := h.Provider.PromoteLoginID(w, r)
			writeResponse(err)
			return err
		}

		return nil
	})
}
