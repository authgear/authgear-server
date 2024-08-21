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
	authInfo *authenticationinfo.T,
	samlSessionEntry *samlsession.SAMLSessionEntry,
) (result httputil.Result) {
	now := h.Clock.NowUTC()
	callbackURL := samlSessionEntry.CallbackURL
	relayState := samlSessionEntry.RelayState

	unexpectedErrorResult := func(err error) httputil.Result {
		return samlprotocolhttp.NewUnexpectedSAMLErrorResult(err,
			samlprotocolhttp.SAMLResult{
				CallbackURL: callbackURL,
				// TODO(saml): Respect the binding protocol set in request
				Binding:    samlprotocol.SAMLBindingHTTPPost,
				Response:   samlprotocol.NewUnexpectedServerErrorResponse(now, h.SAMLService.IdpEntityID()),
				RelayState: relayState,
			},
		)
	}

	defer func() {
		if e := recover(); e != nil {
			// Transform any panic into a saml result
			e := panicutil.MakeError(e)
			result = unexpectedErrorResult(e)
		}
	}()

	var response *samlprotocol.Response
	err := h.Database.WithTx(func() error {
		authnRequest := samlSessionEntry.AuthnRequest()

		resp, err := h.SAMLService.IssueSuccessResponse(
			callbackURL,
			samlSessionEntry.ServiceProviderID,
			*authInfo,
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
			errResponse := samlprotocolhttp.NewExpectedSAMLErrorResult(err,
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
			)
			return errResponse
		}

		return unexpectedErrorResult(err)
	}

	return &samlprotocolhttp.SAMLResult{
		CallbackURL: callbackURL,
		// TODO(saml): Respect the binding protocol set in request
		Binding:    samlprotocol.SAMLBindingHTTPPost,
		Response:   response,
		RelayState: relayState,
	}
}
