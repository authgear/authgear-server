package oauth

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/protocol"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/deps"
)

func AttachRevokeHandler(
	router *mux.Router,
	p *deps.RootProvider,
) {
	router.NewRoute().
		Path("/oauth2/revoke").
		Handler(p.Handler(newRevokeHandler)).
		Methods("POST", "OPTIONS")
}

type oauthRevokeHandler interface {
	Handle(r protocol.RevokeRequest) error
}

type RevokeHandler struct {
	logger        *logrus.Entry
	txContext     db.TxContext
	revokeHandler oauthRevokeHandler
}

func (h *RevokeHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(rw, err.Error(), 400)
		return
	}

	req := protocol.RevokeRequest{}
	for name, values := range r.Form {
		req[name] = values[0]
	}

	err = db.WithTx(h.txContext, func() error {
		return h.revokeHandler.Handle(req)
	})

	if err != nil {
		h.logger.WithError(err).Error("oauth revoke handler failed")
		http.Error(rw, "Internal Server Error", 500)
	}
}
