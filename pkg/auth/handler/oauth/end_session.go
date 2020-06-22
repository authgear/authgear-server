package oauth

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oidc/protocol"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/deps"
)

func AttachEndSessionHandler(
	router *mux.Router,
	p *deps.RootProvider,
) {
	router.NewRoute().
		Path("/oauth2/end_session").
		Handler(p.Handler(newEndSessionHandler)).
		Methods("GET", "POST")
}

type oidcEndSessionHandler interface {
	Handle(auth.AuthSession, protocol.EndSessionRequest, *http.Request, http.ResponseWriter) error
}

type EndSessionHandler struct {
	logger            *logrus.Entry
	txContext         db.TxContext
	endSessionHandler oidcEndSessionHandler
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

	err = db.WithTx(h.txContext, func() error {
		sess := auth.GetSession(r.Context())
		return h.endSessionHandler.Handle(sess, req, r, rw)
	})

	if err != nil {
		h.logger.WithError(err).Error("oauth revoke handler failed")
		http.Error(rw, "Internal Server Error", 500)
	}
}
