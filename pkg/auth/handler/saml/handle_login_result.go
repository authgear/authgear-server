package saml

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlsession"
	"github.com/authgear/authgear-server/pkg/util/clock"
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
) (result SAMLResult) {
	now := h.Clock.NowUTC()
	callbackURL := samlSessionEntry.CallbackURL

	unexpectedErrorResult := func(err error) SAMLResult {
		return NewUnexpectedSAMLErrorResult(err,
			samlprotocol.NewUnexpectedServerErrorResponse(now, h.SAMLService.IdpEntityID()),
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
		authnRequest, _ := samlSessionEntry.AuthnRequest()

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
		var missingNameIDErr *samlprotocol.MissingNameIDError
		if errors.As(err, &missingNameIDErr) {
			errResponse := NewExpectedSAMLErrorResult(err,
				samlprotocol.NewServerErrorResponse(
					now,
					h.SAMLService.IdpEntityID(),
					"missing nameid",
					missingNameIDErr.GetDetailElements(),
				),
			)
			return errResponse
		}

		return unexpectedErrorResult(err)
	}

	return &SAMLSuccessResult{
		Response: response,
	}
}
