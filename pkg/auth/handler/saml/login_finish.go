package saml

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlsession"
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
	SAMLService                HandlerSAMLService
	SAMLSessionService         SAMLSessionService
	AuthenticationInfoResolver SAMLAuthenticationInfoResolver
	AuthenticationInfoService  SAMLAuthenticationInfoService

	LoginResultHandler LoginResultHandler

	BindingHTTPPostWriter BindingHTTPPostWriter
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

	response, err := h.LoginResultHandler.handleLoginResult(&authInfo.T, samlSession.Entry)
	if err != nil {
		h.handleError(rw, r, *samlSession, err)
		return
	}
	err = h.BindingHTTPPostWriter.Write(rw, r,
		samlSession.Entry.CallbackURL,
		response,
		samlSession.Entry.RelayState,
	)
	if err != nil {
		// Don't know how to handle error when writing result, simply panic
		panic(err)
	}
}

func (h *LoginFinishHandler) handleError(
	rw http.ResponseWriter,
	r *http.Request,
	samlSession samlsession.SAMLSession,
	err error,
) {
	now := h.Clock.NowUTC()
	var samlErrResult *SAMLErrorResult
	if errors.As(err, &samlErrResult) {
		if samlErrResult.IsUnexpected {
			h.Logger.WithError(samlErrResult.Cause).Error("unexpected error")
		} else {
			h.Logger.WithError(samlErrResult.Cause).Warnln("saml login failed with expected error")
		}
		err = h.BindingHTTPPostWriter.Write(rw, r,
			samlSession.Entry.CallbackURL,
			samlErrResult.Response,
			samlSession.Entry.RelayState,
		)
		if err != nil {
			panic(err)
		}
	} else {
		h.Logger.WithError(err).Error("unexpected error")
		err = h.BindingHTTPPostWriter.Write(rw, r,
			samlSession.Entry.CallbackURL,
			samlprotocol.NewUnexpectedServerErrorResponse(now, h.SAMLService.IdpEntityID()),
			samlSession.Entry.RelayState,
		)
		if err != nil {
			panic(err)
		}
	}
}
