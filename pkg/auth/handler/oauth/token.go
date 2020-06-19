package oauth

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/handler"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/protocol"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/deps"
)

func AttachTokenHandler(
	router *mux.Router,
	p *deps.RootProvider,
) {
	router.NewRoute().
		Path("/oauth2/token").
		Handler(p.Handler(newTokenHandler)).
		Methods("POST", "OPTIONS")
}

type oauthTokenHandler interface {
	Handle(r protocol.TokenRequest) handler.TokenResult
}

type TokenHandler struct {
	logger       *logrus.Entry
	txContext    db.TxContext
	tokenHandler oauthTokenHandler
}

func (h *TokenHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(rw, err.Error(), 400)
		return
	}

	req := protocol.TokenRequest{}
	for name, values := range r.Form {
		req[name] = values[0]
	}

	var result handler.TokenResult
	err = db.WithTx(h.txContext, func() error {
		result = h.tokenHandler.Handle(req)
		if result.IsInternalError() {
			return errAuthzInternalError
		}
		return nil
	})

	if err == nil || errors.Is(err, errAuthzInternalError) {
		result.WriteResponse(rw, r)
	} else {
		h.logger.WithError(err).Error("oauth token handler failed")
		http.Error(rw, "Internal Server Error", 500)
	}
}
