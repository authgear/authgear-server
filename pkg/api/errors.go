package api

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var (
	InvalidConfiguration = apierrors.InternalError.WithReason("InvalidConfiguration")
	InvalidCredentials   = apierrors.Unauthorized.WithReason("InvalidCredentials")
	InvariantViolated    = apierrors.Invalid.WithReason("InvariantViolated")
)

func NewInvariantViolated(cause string, msg string, data map[string]interface{}) error {
	return InvariantViolated.NewWithCause(
		msg,
		apierrors.MapCause{
			CauseKind: cause,
			Data:      data,
		},
	)
}

var ErrUserNotFound = apierrors.NotFound.WithReason("UserNotFound").New("user not found")
var ErrDuplicatedIdentity = NewInvariantViolated("DuplicatedIdentity", "identity already exists", nil)

var ErrInvalidCredentials = InvalidCredentials.New("invalid credentials")
var ErrOAuthProviderNotFound = apierrors.NotFound.WithReason("OAuthProviderNotFound").New("oauth provider not found")
var ErrIdentityModifyDisabled = NewInvariantViolated("IdentityModifyDisabled", "identity modification disabled", nil)
var ErrMismatchedUser = NewInvariantViolated("MismatchedUser", "mismatched user", nil)
var ErrNoAuthenticator = NewInvariantViolated("NoAuthenticator", "no authenticator", nil)
