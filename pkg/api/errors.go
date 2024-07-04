package api

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var (
	InvariantViolated = apierrors.Invalid.WithReason("InvariantViolated")
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

var UserNotFound = apierrors.NotFound.WithReason("UserNotFound")

var ErrUserNotFound = UserNotFound.New("user not found")
var ErrIdentityNotFound = apierrors.NotFound.WithReason("IdentityNotFound").New("identity not found")

var DuplicatedIdentity = apierrors.AlreadyExists.WithReason("DuplicatedIdentity")
var ErrDuplicatedIdentity = DuplicatedIdentity.New("identity already exist")

const InvalidCredentialsReason = "InvalidCredentials"

var ErrInvalidCredentials = apierrors.Unauthorized.WithReason(InvalidCredentialsReason).New("invalid credentials")

var ErrOAuthProviderNotFound = apierrors.NotFound.WithReason("OAuthProviderNotFound").New("oauth provider not found")
var ErrIdentityModifyDisabled = apierrors.Forbidden.WithReason("IdentityModifyDisabled").New("identity modification disabled")
var ErrMismatchedUser = apierrors.InternalError.WithReason("MismatchedUser").New("mismatched user")
var ErrNoAuthenticator = apierrors.InternalError.WithReason("NoAuthenticator").New("no authenticator")

var ChangePasswordFailed = apierrors.Invalid.WithReason("ChangePasswordFailed")
var ErrNoPassword = ChangePasswordFailed.NewWithCause("the user does not have a password", apierrors.StringCause("NoPassword"))
var ErrPasswordReused = ChangePasswordFailed.NewWithCause("password reused", apierrors.StringCause("PasswordReused"))

const AnonymousUserDisallowedReason = "AnonymousUserDisallowed"

var ErrAnonymousUserDisallowed = apierrors.Invalid.WithReason(AnonymousUserDisallowedReason).New("anonymous user is not enabled in this project")

const BiometricDisallowedReason = "BiometricDisallowed"

var ErrBiometricDisallowed = apierrors.Invalid.WithReason(BiometricDisallowedReason).New("biometric login is not enabled in this project")

const AnonymousUserAddIdentityReason = "AnonymousUserAddIdentity"

var ErrAnonymousUserAddIdentity = apierrors.Invalid.WithReason(AnonymousUserAddIdentityReason).New("anonymous user cannot add identity")

var ErrRemoveLastPrimaryAuthenticator = apierrors.Invalid.WithReason("RemoveLastPrimaryAuthenticator").New("cannot remove the last primary authenticator")
var ErrRemoveLastSecondaryAuthenticator = apierrors.Invalid.WithReason("RemoveLastSecondaryAuthenticator").New("cannot remove the last secondary authenticator")
