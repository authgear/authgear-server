package oauth

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/protocol"
	"github.com/skygeario/skygear-server/pkg/db"
	"github.com/skygeario/skygear-server/pkg/log"
)

func ConfigureRevokeHandler(router *mux.Router, h http.Handler) {
	router.NewRoute().
		Path("/oauth2/revoke").
		Methods("POST", "OPTIONS").
		Handler(h)
}

type RevokeHandlerLogger struct{ *log.Logger }

func NewRevokeHandlerLogger(lf *log.Factory) RevokeHandlerLogger {
	return RevokeHandlerLogger{lf.New("handler-revoke")}
}

type ProtocolRevokeHandler interface {
	Handle(r protocol.RevokeRequest) error
}

type RevokeHandler struct {
	Logger        RevokeHandlerLogger
	DBContext     db.Context
	RevokeHandler ProtocolRevokeHandler
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

	err = db.WithTx(h.DBContext, func() error {
		return h.RevokeHandler.Handle(req)
	})

	if err != nil {
		h.Logger.WithError(err).Error("oauth revoke handler failed")
		http.Error(rw, "Internal Server Error", 500)
	}
}
