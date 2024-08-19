package saml

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol/samlprotocolhttp"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlsession"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/panicutil"
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

	result := h.handleLoginResult(authInfo, samlSession)
	result.WriteResponse(rw, r)
}

func (h *LoginFinishHandler) handleLoginResult(
	authInfo *authenticationinfo.Entry,
	samlSession *samlsession.SAMLSession,
) (result httputil.Result) {
	now := h.Clock.NowUTC()
	callbackURL := samlSession.Entry.CallbackURL
	relayState := samlSession.Entry.RelayState
	defer func() {
		if e := recover(); e != nil {
			e := panicutil.MakeError(e)
			h.Logger.WithError(e).Error("panic")
			result = samlprotocolhttp.NewSAMLErrorResult(e,
				samlprotocolhttp.SAMLResult{
					CallbackURL: callbackURL,
					Response:    samlprotocol.NewInternalServerErrorResponse(now, h.SAMLService.IdpEntityID()),
					RelayState:  relayState,
				},
			)
		}
	}()

	authnRequest := samlSession.Entry.AuthnRequest()
	authenticatedUserID := authInfo.T.UserID

	resp, err := h.SAMLService.IssueSuccessResponse(
		callbackURL,
		samlSession.Entry.ServiceProviderID,
		authenticatedUserID,
		authnRequest,
	)
	if err != nil {
		panic(err)
	}

	return &samlprotocolhttp.SAMLResult{
		CallbackURL: callbackURL,
		// TODO(saml): Respect the binding protocol set in request
		Binding:    samlprotocol.SAMLBindingHTTPPost,
		Response:   resp,
		RelayState: relayState,
	}
}
