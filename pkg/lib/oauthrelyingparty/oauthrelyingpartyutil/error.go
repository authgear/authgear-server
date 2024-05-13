package oauthrelyingpartyutil

import (
	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var InvalidConfiguration = apierrors.InternalError.WithReason("InvalidConfiguration")
var OAuthProtocolError = apierrors.BadRequest.WithReason("OAuthProtocolError")
var OAuthError = apierrors.BadRequest.WithReason("OAuthError")

func NewOAuthError(errorString string, errorDescription string, errorURI string) error {
	msg := errorString
	if errorDescription != "" {
		msg += ": " + errorDescription
	}

	return OAuthError.NewWithInfo(msg, apierrors.Details{
		"error":             errorString,
		"error_description": errorDescription,
		"error_uri":         errorURI,
	})
}

func ErrorResponseAsError(errResp oauthrelyingparty.ErrorResponse) error {
	return NewOAuthError(errResp.Error, errResp.ErrorDescription, errResp.ErrorURI)
}
