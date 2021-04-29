package sso

import (
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

// oauthErrorResp is a helper struct for deserialization purpose.
type oauthErrorResp struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
}

func (r *oauthErrorResp) AsError() error {
	return NewOAuthError(r.Error, r.ErrorDescription, r.ErrorURI)
}
