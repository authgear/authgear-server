package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

func AttachSignupOOBOTPHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.
		NewRoute().
		Path("/signup/oob_otp").
		Methods("OPTIONS", "POST", "GET").
		Handler(auth.MakeHandler(authDependency, newSignupOOBOTPHandler))
}

type signupOOBOTPProvider interface {
	GetSignupOOBOTPForm(w http.ResponseWriter, r *http.Request) (func(err error), error)
	PostSignupOOBOTP(w http.ResponseWriter, r *http.Request) (func(err error), error)
	TriggerSignupOOBOTP(w http.ResponseWriter, r *http.Request) (func(err error), error)
}

type SignupOOBOTPHandler struct {
	Provider  signupOOBOTPProvider
	TxContext db.TxContext
}

func (h *SignupOOBOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db.WithTx(h.TxContext, func() error {
		if r.Method == "GET" {
			writeResponse, err := h.Provider.GetSignupOOBOTPForm(w, r)
			writeResponse(err)
			return err
		}

		if r.Method == "POST" {
			if r.Form.Get("trigger") == "true" {
				r.Form.Del("trigger")
				writeResponse, err := h.Provider.TriggerSignupOOBOTP(w, r)
				writeResponse(err)
				return err
			}

			writeResponse, err := h.Provider.PostSignupOOBOTP(w, r)
			writeResponse(err)
			return err
		}

		return nil
	})
}
