package oauthrelyingpartyutil

import (
	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
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
