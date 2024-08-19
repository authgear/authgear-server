package saml

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func ConfigureLoginFinishRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET").
		WithPathPattern("/saml2/login_finish")
}

type LoginFinishHandlerLogger struct{ *log.Logger }

func NewLoginFinishHandlerLogger(lf *log.Factory) *LoginFinishHandlerLogger {
	return &LoginFinishHandlerLogger{lf.New("saml-login-finish-handler")}
}

type LoginFinishHandler struct {
	Logger *LoginFinishHandlerLogger
}

func (h *LoginFinishHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

}
