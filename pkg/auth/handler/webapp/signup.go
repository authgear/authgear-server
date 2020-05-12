package webapp

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

func AttachSignupHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.
		NewRoute().
		Path("/signup").
		Methods("OPTIONS", "POST", "GET").
		Handler(auth.MakeHandler(authDependency, newSignupHandler))
}

type signupProvider interface {
	GetCreateLoginIDForm(w http.ResponseWriter, r *http.Request) (func(error), error)
	CreateLoginID(w http.ResponseWriter, r *http.Request) (func(error), error)
	ChooseIdentityProvider(w http.ResponseWriter, r *http.Request, oauthProvider webapp.OAuthProvider) (func(error), error)
}

type SignupHandler struct {
	Provider      signupProvider
	oauthProvider webapp.OAuthProvider
	TxContext     db.TxContext
}

func (h *SignupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db.WithTx(h.TxContext, func() error {
		if r.Method == "GET" {
			writeResponse, err := h.Provider.GetCreateLoginIDForm(w, r)
			writeResponse(err)
			return err
		}

		if r.Method == "POST" {
			if r.Form.Get("x_idp_id") != "" {
				if h.oauthProvider == nil {
					http.Error(w, "Not found", http.StatusNotFound)
					return errors.New("oauth provider not found")
				}
				writeResponse, err := h.Provider.ChooseIdentityProvider(w, r, h.oauthProvider)
				writeResponse(err)
				return err
			}

			writeResponse, err := h.Provider.CreateLoginID(w, r)
			writeResponse(err)
			return err
		}

		return nil
	})
}
