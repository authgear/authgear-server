package oauth

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oidc/protocol"
	"github.com/skygeario/skygear-server/pkg/db"
	"github.com/skygeario/skygear-server/pkg/log"
)

func ConfigureEndSessionHandler(router *mux.Router, h http.Handler) {
	router.NewRoute().
		Path("/oauth2/end_session").
		Methods("GET", "POST").
		Handler(h)
}

type oidcEndSessionHandler interface {
	Handle(auth.AuthSession, protocol.EndSessionRequest, *http.Request, http.ResponseWriter) error
}

type EndSessionHandler struct {
	Logger            *log.Logger
	DBContext         db.Context
	EndSessionHandler oidcEndSessionHandler
}

func (h *EndSessionHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(rw, err.Error(), 400)
		return
	}

	req := protocol.EndSessionRequest{}
	for name, values := range r.Form {
		req[name] = values[0]
	}

	err = db.WithTx(h.DBContext, func() error {
		sess := auth.GetSession(r.Context())
		return h.EndSessionHandler.Handle(sess, req, r, rw)
	})

	if err != nil {
		h.Logger.WithError(err).Error("oauth revoke handler failed")
		http.Error(rw, "Internal Server Error", 500)
	}
}
