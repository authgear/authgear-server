package oauth

import (
	"context"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

func ConfigureTokenRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("POST", "OPTIONS").
		WithPathPattern("/oauth2/token")
}

type ProtocolTokenHandler interface {
	Handle(ctx context.Context, rw http.ResponseWriter, req *http.Request, r protocol.TokenRequest) httputil.Result
}

var TokenHandlerLogger = slogutil.NewLogger("handler-token")

type TokenHandler struct {
	Database     *appdb.Handle
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
		req[name] = values
	}

	var result httputil.Result
	ctx := r.Context()
	result = h.TokenHandler.Handle(ctx, rw, r, req)
	if result.IsInternalError() {
		err = errAuthzInternalError
	}

	if err == nil || errors.Is(err, errAuthzInternalError) {
		result.WriteResponse(rw, r)
	} else {
		logger := TokenHandlerLogger.GetLogger(ctx)
		logger.WithError(err).Error(ctx, "oauth token handler failed")
		http.Error(rw, "Internal Server Error", 500)
	}
}
