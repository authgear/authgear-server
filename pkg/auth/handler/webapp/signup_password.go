package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

func AttachSignupPasswordHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.
		NewRoute().
		Path("/signup/password").
		Methods("OPTIONS", "POST", "GET").
		Handler(auth.MakeHandler(authDependency, newSignupPasswordHandler))
}

type signupPasswordProvider interface {
	GetSignupPasswordForm(w http.ResponseWriter, r *http.Request) (func(err error), error)
	PostSignupPassword(w http.ResponseWriter, r *http.Request) (func(err error), error)
}

type SignupPasswordHandler struct {
	Provider  signupPasswordProvider
	TxContext db.TxContext
}

func (h *SignupPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db.WithTx(h.TxContext, func() error {
		if r.Method == "GET" {
			writeResponse, err := h.Provider.GetSignupPasswordForm(w, r)
			writeResponse(err)
			return err
		}

		if r.Method == "POST" {
			writeResponse, err := h.Provider.PostSignupPassword(w, r)
			writeResponse(err)
			return err
		}

		return nil
	})
}
