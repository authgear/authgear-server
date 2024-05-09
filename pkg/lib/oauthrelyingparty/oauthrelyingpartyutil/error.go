package oauthrelyingpartyutil

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
)

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
