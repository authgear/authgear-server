package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

func AttachEnterLoginIDHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.
		NewRoute().
		Path("/enter_login_id").
		Methods("OPTIONS", "POST", "GET").
		Handler(auth.MakeHandler(authDependency, newEnterLoginIDHandler))
}

type EnterLoginIDProvider interface {
	GetEnterLoginIDForm(w http.ResponseWriter, r *http.Request) (func(error), error)
	EnterLoginID(w http.ResponseWriter, r *http.Request) (func(error), error)
}

type EnterLoginIDHandler struct {
	Provider  EnterLoginIDProvider
	TxContext db.TxContext
}

func (h *EnterLoginIDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db.WithTx(h.TxContext, func() error {
		if r.Method == "GET" {
			writeResponse, err := h.Provider.GetEnterLoginIDForm(w, r)
			writeResponse(err)
			return err
		}

		if r.Method == "POST" {
			writeResponse, err := h.Provider.EnterLoginID(w, r)
			writeResponse(err)
			return err
		}

		return nil
	})
}
