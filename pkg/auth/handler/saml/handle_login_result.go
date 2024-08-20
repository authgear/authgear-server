package saml

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlerror"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol/samlprotocolhttp"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlsession"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/panicutil"
)

type LoginResultHandler struct {
	Clock       clock.Clock
	Database    *appdb.Handle
	SAMLService HandlerSAMLService
}

func (h *LoginResultHandler) handleLoginResult(
	authInfo *authenticationinfo.Entry,
	samlSession *samlsession.SAMLSession,
) (result httputil.Result) {
	now := h.Clock.NowUTC()
	callbackURL := samlSession.Entry.CallbackURL
	relayState := samlSession.Entry.RelayState
	authenticatedUserID := authInfo.T.UserID
	defer func() {
		if e := recover(); e != nil {
			e := panicutil.MakeError(e)
			result = samlprotocolhttp.NewSAMLErrorResult(e,
				samlprotocolhttp.SAMLResult{
					CallbackURL: callbackURL,
					// TODO(saml): Respect the binding protocol set in request
					Binding:    samlprotocol.SAMLBindingHTTPPost,
					Response:   samlprotocol.NewUnexpectedServerErrorResponse(now, h.SAMLService.IdpEntityID()),
					RelayState: relayState,
				},
				true,
			)
		}
	}()

	var response *samlprotocol.Response
	err := h.Database.WithTx(func() error {
		authnRequest := samlSession.Entry.AuthnRequest()

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
				},
				false,
			)
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
