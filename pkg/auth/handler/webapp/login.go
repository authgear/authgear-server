package webapp

import (
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

func AttachLoginHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.
		NewRoute().
		Path("/login").
		Methods("OPTIONS", "POST", "GET").
		Handler(auth.MakeHandler(authDependency, newLoginHandler))
}

func redirectURIForWebApp(urlPrefix *url.URL, providerConfig config.OAuthProviderConfiguration) string {
	u := *urlPrefix
	u.Path = path.Join(u.Path, fmt.Sprintf("sso/oauth2/callback/%s", url.PathEscape(providerConfig.ID)))
	return u.String()
}

type loginProvider interface {
	GetLoginForm(w http.ResponseWriter, r *http.Request) (func(error), error)
	EnterLoginID(w http.ResponseWriter, r *http.Request) (func(error), error)
	LoginIdentityProvider(w http.ResponseWriter, r *http.Request, providerAlias string) (func(error), error)
}

type LoginHandler struct {
	Provider  loginProvider
	TxContext db.TxContext
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db.WithTx(h.TxContext, func() error {
		if r.Method == "GET" {
			writeResponse, err := h.Provider.GetLoginForm(w, r)
			writeResponse(err)
			return err
		}

		if r.Method == "POST" {
			if r.Form.Get("x_idp_id") != "" {
				writeResponse, err := h.Provider.LoginIdentityProvider(w, r, r.Form.Get("x_idp_id"))
				writeResponse(err)
				return err
			}

			writeResponse, err := h.Provider.EnterLoginID(w, r)
			writeResponse(err)
			return err
		}

		return nil
	})

	return
}
