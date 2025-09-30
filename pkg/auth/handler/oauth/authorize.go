package oauth

import (
	"context"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

func ConfigureAuthorizeRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "POST").
		WithPathPattern("/oauth2/authorize")
}

var AuthorizeHandlerLogger = slogutil.NewLogger("handler-authz")

type ProtocolAuthorizeHandler interface {
	ValidateRequestWithoutTx(ctx context.Context, r protocol.AuthorizationRequest) (context.Context, *handler.AuthorizationParams, *handler.AuthorizationResultError)
	HandleRequest(ctx context.Context, r protocol.AuthorizationRequest, params *handler.AuthorizationParams) httputil.Result
}

var errAuthzInternalError = errors.New("internal error")

type AuthorizeHandler struct {
	AuthzHandler ProtocolAuthorizeHandler
}

func (h *AuthorizeHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(rw, err.Error(), 400)
		return
	}

	req := protocol.AuthorizationRequest{}
	for name, values := range r.Form {
		req[name] = values[0]
	}

	ctx, params, errResult := h.AuthzHandler.ValidateRequestWithoutTx(r.Context(), req)
	if errResult != nil {
		errResult.WriteResponse(rw, r)
		return
	}

	var result httputil.Result
	result = h.AuthzHandler.HandleRequest(ctx, req, params)
	if result.IsInternalError() {
		err = errAuthzInternalError
	}
	if err == nil || errors.Is(err, errAuthzInternalError) {
		result.WriteResponse(rw, r)
	} else {
		logger := AuthorizeHandlerLogger.GetLogger(r.Context())
		logger.WithError(err).Error(r.Context(), "oauth authz handler failed")
		http.Error(rw, "Internal Server Error", 500)
	}
}
