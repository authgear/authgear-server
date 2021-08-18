package oauth

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func ConfigureFromWebAppRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "POST").
		WithPathPattern("/oauth2/_from_webapp")
}

type FromWebAppHandlerLogger struct{ *log.Logger }

func NewFromWebAppHandlerLogger(lf *log.Factory) FromWebAppHandlerLogger {
	return FromWebAppHandlerLogger{lf.New("handler-from-webapp")}
}

type ProtocolFromWebAppHandler interface {
	HandleFromWebApp(r protocol.AuthorizationRequest, req *http.Request) httputil.Result
}

type FromWebAppHandler struct {
	Logger   FromWebAppHandlerLogger
	Database *appdb.Handle
	Handler  ProtocolFromWebAppHandler
}

func (h *FromWebAppHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(rw, err.Error(), 400)
		return
	}

	req := protocol.AuthorizationRequest{}
	for name, values := range r.Form {
		req[name] = values[0]
	}

	var result httputil.Result
	err = h.Database.WithTx(func() error {
		result = h.Handler.HandleFromWebApp(req, r)
		if result.IsInternalError() {
			return errAuthzInternalError
		}
		return nil
	})

	if err == nil || errors.Is(err, errAuthzInternalError) {
		result.WriteResponse(rw, r)
	} else {
		h.Logger.WithError(err).Error("")
		http.Error(rw, "Internal Server Error", 500)
	}
}
