package saml

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
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
	ResultWriter       SAMLResultWriter
}

func (h *LoginFinishHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	authInfoID, ok := h.AuthenticationInfoResolver.GetAuthenticationInfoID(r)
	if !ok {
		h.Logger.Warningln("authentication info id is missing")
		// Maybe the user visited the page directly, tell him not to do so.
		http.Error(rw, "invoking this endpoint directly is not supported", http.StatusBadRequest)
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

	result := h.LoginResultHandler.handleLoginResult(&authInfo.T, samlSession.Entry)
	switch result := result.(type) {
	case *samlprotocolhttp.SAMLErrorResult:
		if result.IsUnexpected {
			h.Logger.WithError(result.Cause).Error("unexpected error")
		} else {
			h.Logger.WithError(result.Cause).Warnln("saml login failed with expected error")
		}
	}
	err = h.ResultWriter.Write(rw, r, result, &samlprotocolhttp.WriteOptions{
		Binding:     samlprotocol.SAMLBindingHTTPPost,
		CallbackURL: samlSession.Entry.CallbackURL,
		RelayState:  samlSession.Entry.RelayState,
	})
	if err != nil {
		// Don't know how to handle error when writing result, simply panic
		panic(err)
	}
}
