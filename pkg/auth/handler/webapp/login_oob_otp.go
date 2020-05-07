package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

func AttachLoginOOBOTPHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.
		NewRoute().
		Path("/login/oob_otp").
		Methods("OPTIONS", "POST", "GET").
		Handler(auth.MakeHandler(authDependency, newLoginOOBOTPHandler))
}

type loginOOBOTPProvider interface {
	GetOOBOTPForm(w http.ResponseWriter, r *http.Request) (func(err error), error)
	EnterSecret(w http.ResponseWriter, r *http.Request) (func(err error), error)
	TriggerOOBOTP(w http.ResponseWriter, r *http.Request) (func(err error), error)
}

type LoginOOBOTPHandler struct {
	Provider  loginOOBOTPProvider
	TxContext db.TxContext
}

func (h *LoginOOBOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db.WithTx(h.TxContext, func() error {
		if r.Method == "GET" {
			writeResponse, err := h.Provider.GetOOBOTPForm(w, r)
			writeResponse(err)
			return err
		}

		if r.Method == "POST" {
			if r.Form.Get("trigger") == "true" {
				r.Form.Del("trigger")
				writeResponse, err := h.Provider.TriggerOOBOTP(w, r)
				writeResponse(err)
				return err
			}

			writeResponse, err := h.Provider.EnterSecret(w, r)
			writeResponse(err)
			return err
		}

		return nil
	})
}
