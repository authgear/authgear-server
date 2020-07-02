package oauth

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/dependency/oauth/handler"
	"github.com/authgear/authgear-server/pkg/auth/dependency/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/log"
)

func ConfigureTokenRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("POST", "OPTIONS").
		WithPathPattern("/oauth2/token")
}

type ProtocolTokenHandler interface {
	Handle(r protocol.TokenRequest) handler.TokenResult
}

type TokenHandlerLogger struct{ *log.Logger }

func NewTokenHandlerLogger(lf *log.Factory) TokenHandlerLogger {
	return TokenHandlerLogger{lf.New("handler-token")}
}

type TokenHandler struct {
	Logger       TokenHandlerLogger
	DBContext    db.Context
	TokenHandler ProtocolTokenHandler
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
	err = db.WithTx(h.DBContext, func() error {
		result = h.TokenHandler.Handle(req)
		if result.IsInternalError() {
			return errAuthzInternalError
		}
		return nil
	})

	if err == nil || errors.Is(err, errAuthzInternalError) {
		result.WriteResponse(rw, r)
	} else {
		h.Logger.WithError(err).Error("oauth token handler failed")
		http.Error(rw, "Internal Server Error", 500)
	}
}
