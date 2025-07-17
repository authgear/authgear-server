package oauth

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

func ConfigureEndSessionRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "POST").
		WithPathPattern("/oauth2/end_session")
}

var EndSessionHandlerLogger = slogutil.NewLogger("handler-end-session")

type ProtocolEndSessionHandler interface {
	Handle(ctx context.Context, s session.ResolvedSession, endSessionRequest protocol.EndSessionRequest, r *http.Request, w http.ResponseWriter) error
}

type EndSessionHandler struct {
	Database          *appdb.Handle
	EndSessionHandler ProtocolEndSessionHandler
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

	err = h.Database.WithTx(r.Context(), func(ctx context.Context) error {
		sess := session.GetSession(ctx)
		return h.EndSessionHandler.Handle(ctx, sess, req, r, rw)
	})

	if err != nil {
		logger := EndSessionHandlerLogger.GetLogger(r.Context())
		logger.WithError(err).Error(r.Context(), "oauth revoke handler failed")
		http.Error(rw, "Internal Server Error", 500)
	}
}
