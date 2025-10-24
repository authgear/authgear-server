package oauthrelyingpartyutil

import (
	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var InvalidConfiguration = apierrors.InternalError.WithReason("InvalidConfiguration")
var OAuthProtocolError = apierrors.BadRequest.WithReason("OAuthProtocolError")
var OAuthError = apierrors.BadRequest.WithReason("OAuthError")

func NewOAuthError(errResp *oauthrelyingparty.ErrorResponse) error {
	return OAuthError.NewWithInfo(errResp.Error(), apierrors.Details{
		"error":             errResp.Error_,
		"error_description": errResp.ErrorDescription,
		"error_uri":         errResp.ErrorURI,
	})
}

type OAuthRelyingPartyInternalError struct {
	err                error
	IsLoggingSkippable bool
}

var _ error = (*OAuthRelyingPartyInternalError)(nil)
var _ slogutil.LoggingSkippable = (*OAuthRelyingPartyInternalError)(nil)

func (e *OAuthRelyingPartyInternalError) Error() string {
	return e.err.Error()
}

func (e *OAuthRelyingPartyInternalError) SkipLogging() bool {
	return e.IsLoggingSkippable
}
