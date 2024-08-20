package saml

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol/samlprotocolhttp"
	"github.com/authgear/authgear-server/pkg/util/clock"
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
	Logger                     *LoginFinishHandlerLogger
	Clock                      clock.Clock
	SAMLSessionService         SAMLSessionService
	AuthenticationInfoResolver SAMLAuthenticationInfoResolver
	AuthenticationInfoService  SAMLAuthenticationInfoService

	LoginResultHandler LoginResultHandler
}

func (h *LoginFinishHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	authInfoID, ok := h.AuthenticationInfoResolver.GetAuthenticationInfoID(r)
	if !ok {
		h.Logger.Warningln("authentication info id is missing")
		http.NotFound(rw, r)
		return
	}

	authInfo, err := h.AuthenticationInfoService.Get(authInfoID)
	if err != nil {
		// It is unexpected that we've set the code in query, but turns out it does not exist.
		// We have no idea how to return the error response to user yet,
		// so we can only panic.
		panic(err)
	}

	samlSession, err := h.SAMLSessionService.Get(authInfo.SAMLSessionID)
	if err != nil {
		// It is unexpected that we've set the session id in auth info, but turns out it does not exist.
		// We have no idea how to return the error response to user yet,
		// so we can only panic.
		panic(err)
	}

	result := h.LoginResultHandler.handleLoginResult(authInfo, samlSession)
	switch result := result.(type) {
	case *samlprotocolhttp.SAMLErrorResult:
		if result.IsUnexpected {
			h.Logger.WithError(result.Cause).Error("unexpected error")
		}
	}
	result.WriteResponse(rw, r)
}
