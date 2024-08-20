package saml

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlerror"
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
	Database                   *appdb.Handle
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
					// TODO(saml): Respect the binding protocol set in request
					Binding:    samlprotocol.SAMLBindingHTTPPost,
					Response:   samlprotocol.NewUnexpectedServerErrorResponse(now, h.SAMLService.IdpEntityID()),
					RelayState: relayState,
				},
			)
		}
	}()

	var response *samlprotocol.Response
	err := h.Database.WithTx(func() error {
		authnRequest := samlSession.Entry.AuthnRequest()
		authenticatedUserID := authInfo.T.UserID

		resp, err := h.SAMLService.IssueSuccessResponse(
			callbackURL,
			samlSession.Entry.ServiceProviderID,
			authenticatedUserID,
			authnRequest,
		)
		if err != nil {
			return err
		}
		response = resp
		return nil
	})
	if err != nil {
		var missingNameIDErr *samlerror.MissingNameIDError
		if errors.As(err, &missingNameIDErr) {
			errResponse := samlprotocolhttp.NewSAMLErrorResult(err,
				samlprotocolhttp.SAMLResult{
					CallbackURL: callbackURL,
					// TODO(saml): Respect the binding protocol set in request
					Binding: samlprotocol.SAMLBindingHTTPPost,
					Response: samlprotocol.NewServerErrorResponse(
						now,
						h.SAMLService.IdpEntityID(),
						"missing nameid",
						missingNameIDErr.GetDetailElements(),
					),
					RelayState: relayState,
				})
			return errResponse
		}
		panic(err)
	}

	return &samlprotocolhttp.SAMLResult{
		CallbackURL: callbackURL,
		// TODO(saml): Respect the binding protocol set in request
		Binding:    samlprotocol.SAMLBindingHTTPPost,
		Response:   response,
		RelayState: relayState,
	}
}
