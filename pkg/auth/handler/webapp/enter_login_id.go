package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/db"
	"github.com/skygeario/skygear-server/pkg/deps"
)

func AttachEnterLoginIDHandler(
	router *mux.Router,
	p *deps.RootProvider,
) {
	router.
		NewRoute().
		Path("/enter_login_id").
		Methods("OPTIONS", "POST", "GET").
		Handler(p.Handler(newEnterLoginIDHandler))
}

type EnterLoginIDProvider interface {
	GetEnterLoginIDForm(w http.ResponseWriter, r *http.Request) (func(error), error)
	EnterLoginID(w http.ResponseWriter, r *http.Request) (func(error), error)
	RemoveLoginID(w http.ResponseWriter, r *http.Request) (func(error), error)
}

type EnterLoginIDHandler struct {
	Provider  EnterLoginIDProvider
	DBContext db.Context
}

func (h *EnterLoginIDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db.WithTx(h.DBContext, func() error {
		if r.Method == "GET" {
			writeResponse, err := h.Provider.GetEnterLoginIDForm(w, r)
			writeResponse(err)
			return err
		}

		if r.Method == "POST" {
			if r.Form.Get("x_action") == "remove" {
				writeResponse, err := h.Provider.RemoveLoginID(w, r)
				writeResponse(err)
				return err
			}

			writeResponse, err := h.Provider.EnterLoginID(w, r)
			writeResponse(err)
			return err
		}

		return nil
	})
}
