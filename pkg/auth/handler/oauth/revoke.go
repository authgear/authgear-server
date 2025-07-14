package oauth

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

func ConfigureRevokeRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("POST", "OPTIONS").
		WithPathPattern("/oauth2/revoke")
}

var RevokeHandlerLogger = slogutil.NewLogger("handler-revoke")

type ProtocolRevokeHandler interface {
	Handle(ctx context.Context, r protocol.RevokeRequest) error
}

type RevokeHandler struct {
	Database      *appdb.Handle
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

	ctx := r.Context()
	err = h.Database.WithTx(ctx, func(ctx context.Context) error {
		return h.RevokeHandler.Handle(ctx, req)
	})

	if err != nil {
		logger := RevokeHandlerLogger.GetLogger(ctx)
		logger.WithError(err).Error(ctx, "oauth revoke handler failed")
		http.Error(rw, "Internal Server Error", 500)
	}
}
