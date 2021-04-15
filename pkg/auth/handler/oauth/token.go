package oauth

import (
	"errors"
	"net/http"

	tenantdb "github.com/authgear/authgear-server/pkg/lib/infra/db/tenant"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func ConfigureTokenRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("POST", "OPTIONS").
		WithPathPattern("/oauth2/token")
}

type ProtocolTokenHandler interface {
	Handle(rw http.ResponseWriter, req *http.Request, r protocol.TokenRequest) httputil.Result
}

type TokenHandlerLogger struct{ *log.Logger }

func NewTokenHandlerLogger(lf *log.Factory) TokenHandlerLogger {
	return TokenHandlerLogger{lf.New("handler-token")}
}

type TokenHandler struct {
	Logger       TokenHandlerLogger
	Database     *tenantdb.Handle
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

	var result httputil.Result
	err = h.Database.WithTx(func() error {
		result = h.TokenHandler.Handle(rw, r, req)
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
