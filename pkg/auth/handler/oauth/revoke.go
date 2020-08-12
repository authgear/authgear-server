package oauth

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/dependency/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func ConfigureRevokeRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("POST", "OPTIONS").
		WithPathPattern("/oauth2/revoke")
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
	Database      *db.Handle
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

	err = h.Database.WithTx(func() error {
		return h.RevokeHandler.Handle(req)
	})

	if err != nil {
		h.Logger.WithError(err).Error("oauth revoke handler failed")
		http.Error(rw, "Internal Server Error", 500)
	}
}
