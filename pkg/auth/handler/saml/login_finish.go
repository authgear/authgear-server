package saml

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlsession"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/panicutil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

func ConfigureLoginFinishRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET").
		WithPathPattern("/saml2/login_finish")
}

var LoginFinishHandlerLogger = slogutil.NewLogger("saml-login-finish-handler")

type LoginFinishHandler struct {
	Clock                      clock.Clock
	SAMLService                HandlerSAMLService
	SAMLSessionService         SAMLSessionService
	AuthenticationInfoResolver SAMLAuthenticationInfoResolver
	AuthenticationInfoService  SAMLAuthenticationInfoService

	LoginResultHandler LoginResultHandler

	BindingHTTPPostWriter BindingHTTPPostWriter
}

func (h *LoginFinishHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := LoginFinishHandlerLogger.GetLogger(ctx)

	authInfoID, ok := h.AuthenticationInfoResolver.GetAuthenticationInfoID(r)
	if !ok {
		logger.Warn(ctx, "authentication info id is missing")
		// Maybe the user visited the page directly, tell him not to do so.
		http.Error(rw, "invoking this endpoint directly is not supported", http.StatusBadRequest)
		return
	}

	authInfo, err := h.AuthenticationInfoService.Get(ctx, authInfoID)
	if err != nil {
		// It is unexpected that we've set the code in query, but turns out it does not exist.
		// We have no idea how to return the error response to user yet,
		// so we can only panic.
		panic(err)
	}

	samlSession, err := h.SAMLSessionService.Get(ctx, authInfo.SAMLSessionID)
	if err != nil {
		// It is unexpected that we've set the session id in auth info, but turns out it does not exist.
		// We have no idea how to return the error response to user yet,
		// so we can only panic.
		panic(err)
	}

	defer func() {
		if err := recover(); err != nil {
			e := panicutil.MakeError(err)
			h.handleError(rw, r,
				samlSession,
				e,
			)
		}
	}()

	response, err := h.LoginResultHandler.handleLoginResult(
		ctx,
		&authInfo.T,
		samlSession.Entry,
	)
	if err != nil {
		h.handleError(rw, r, samlSession, err)
		return
	}
	err = h.BindingHTTPPostWriter.WriteResponse(rw, r,
		samlSession.Entry.CallbackURL,
		response.Element(),
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
	samlSession *samlsession.SAMLSession,
	err error,
) {
	ctx := r.Context()
	logger := LoginFinishHandlerLogger.GetLogger(ctx)
	now := h.Clock.NowUTC()
	var samlErrResult *SAMLErrorResult
	if errors.As(err, &samlErrResult) {
		logger.WithError(samlErrResult.Cause).Warn(r.Context(), "saml login failed with expected error")
		err = h.BindingHTTPPostWriter.WriteResponse(rw, r,
			samlSession.Entry.CallbackURL,
			samlErrResult.Response.Element(),
			samlSession.Entry.RelayState,
		)
		if err != nil {
			panic(err)
		}
	} else {
		logger.WithError(err).Error(r.Context(), "unexpected error")
		err = h.BindingHTTPPostWriter.WriteResponse(rw, r,
			samlSession.Entry.CallbackURL,
			samlprotocol.NewUnexpectedServerErrorResponse(now, h.SAMLService.IdpEntityID()).Element(),
			samlSession.Entry.RelayState,
		)
		if err != nil {
			panic(err)
		}
	}
}
