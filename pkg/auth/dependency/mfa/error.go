package mfa

import (
	skyerr "github.com/skygeario/skygear-server/pkg/core/xskyerr"
)

var (
	InvalidRecoveryCode        = skyerr.Unauthorized.WithReason("InvalidRecoveryCode")
	InvalidBearerToken         = skyerr.Unauthorized.WithReason("InvalidMFABearerToken")
	InvalidMFACode             = skyerr.Unauthorized.WithReason("InvalidMFACode")
	AuthenticatorNotFound      = skyerr.NotFound.WithReason("AuthenticatorNotFound")
	AuthenticatorAlreadyExists = skyerr.AlreadyExists.WithReason("AuthenticatorAlreadyExists")
	InvalidMFARequest          = skyerr.Invalid.WithReason("InvalidMFARequest")
)

var (
	errInvalidRecoveryCode        = InvalidRecoveryCode.New("invalid recovery code")
	errInvalidBearerToken         = InvalidBearerToken.New("invalid bearer token")
	errInvalidMFACode             = InvalidMFACode.New("invalid MFA code")
	errAuthenticatorNotFound      = AuthenticatorNotFound.New("authenticator not found")
	errAuthenticatorAlreadyExists = AuthenticatorAlreadyExists.New("authenticator already exists")
)

type invalidMFARequestCause string

const (
	TooManyAuthenticator  invalidMFARequestCause = "TooManyAuthenticator"
	IncorrectCode         invalidMFARequestCause = "IncorrectCode"
	AuthenticatorRequired invalidMFARequestCause = "AuthenticatorRequired"
)

func NewInvalidMFARequest(cause invalidMFARequestCause, msg string) error {
	return InvalidMFARequest.NewWithDetails(msg, skyerr.Details{"cause": skyerr.APIErrorString(cause)})
}
