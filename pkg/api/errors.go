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

var UserNotFound = apierrors.NotFound.WithReason("UserNotFound")

var ErrUserNotFound = UserNotFound.New("user not found")
var ErrIdentityNotFound = apierrors.NotFound.WithReason("IdentityNotFound").New("identity not found")

var ErrInvalidCredentials = InvalidCredentials.New("invalid credentials")
var ErrOAuthProviderNotFound = apierrors.NotFound.WithReason("OAuthProviderNotFound").New("oauth provider not found")
var ErrIdentityModifyDisabled = NewInvariantViolated("IdentityModifyDisabled", "identity modification disabled", nil)
var ErrMismatchedUser = NewInvariantViolated("MismatchedUser", "mismatched user", nil)
var ErrNoAuthenticator = NewInvariantViolated("NoAuthenticator", "no authenticator", nil)
var ErrClaimNotVerifiable = NewInvariantViolated("ClaimNotVerifiable", "claim not verifiable", nil)

var ChangePasswordFailed = apierrors.Invalid.WithReason("ChangePasswordFailed")
var ErrNoPassword = ChangePasswordFailed.NewWithCause("the user does not have a password", apierrors.StringCause("NoPassword"))
var ErrPasswordReused = ChangePasswordFailed.NewWithCause("password reused", apierrors.StringCause("PasswordReused"))

var ErrLDAPServerNotFound = apierrors.NotFound.WithReason("LDAPServerNotFound").New("ldap server not found")
var LDAPConnectionTestFailed = apierrors.ServiceUnavailable.WithReason("LDAPConnectionTestFailed")
var ErrLDAPCannotConnect = LDAPConnectionTestFailed.NewWithCause("failed to connect", apierrors.StringCause("FailedToConnect"))
var ErrLDAPFailedToBindSearchUser = LDAPConnectionTestFailed.NewWithCause("failed to bind search user", apierrors.StringCause("FailedToBindSearchUser"))
var ErrLDAPEndUserSearchNotFound = LDAPConnectionTestFailed.NewWithCause("end user not found", apierrors.StringCause("TestingEndUserNotFound"))
var ErrLDAPEndUserSearchMultipleResult = LDAPConnectionTestFailed.NewWithCause("multiple end users found", apierrors.StringCause("MoreThanOneEntryInSearchResult"))
var ErrLDAPMissingUniqueAttribute = LDAPConnectionTestFailed.NewWithCause("missing ID attribute", apierrors.StringCause("TestingEndUserMissingUserIDAttribute"))
