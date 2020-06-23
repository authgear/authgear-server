package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/db"
)

func ConfigureCreatePasswordHandler(router *mux.Router, h http.Handler) {
	router.NewRoute().
		Path("/create_password").
		Methods("OPTIONS", "POST", "GET").
		Handler(h)
}

type createPasswordProvider interface {
	GetCreatePasswordForm(w http.ResponseWriter, r *http.Request) (func(err error), error)
	EnterSecret(w http.ResponseWriter, r *http.Request) (func(err error), error)
}

type CreatePasswordHandler struct {
	Provider  createPasswordProvider
	DBContext db.Context
}

func (h *CreatePasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db.WithTx(h.DBContext, func() error {
		if r.Method == "GET" {
			writeResponse, err := h.Provider.GetCreatePasswordForm(w, r)
			writeResponse(err)
			return err
		}

		if r.Method == "POST" {
			writeResponse, err := h.Provider.EnterSecret(w, r)
			writeResponse(err)
			return err
		}

		return nil
	})
}
