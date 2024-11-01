package accountmanagement

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrOAuthTokenInvalid = apierrors.Invalid.WithReason("AccountManagementOAuthTokenInvalid").New("invalid token")
var ErrOAuthStateNotBoundToToken = apierrors.Invalid.WithReason("AccountManagementOAuthStateNotBoundToToken").New("the state parameter in query is not bound to token")
var ErrOAuthTokenNotBoundToUser = apierrors.Invalid.WithReason("AccountManagementOAuthTokenNotBoundToUser").New("token is not bound to the current user")

var ErrAccountManagementTokenInvalid = apierrors.Invalid.WithReason("AccountManagementTokenInvalid").New("invalid token")
var ErrAccountManagementTokenNotBoundToUser = apierrors.Invalid.WithReason("AccountManagementTokenNotBoundToUser").New("token is not bound to the current user")

var ErrAccountManagementIdentityNotOwnedbyToUser = apierrors.Invalid.WithReason("AccountManagementIdentityNotOwnedByUser").New("identity not owned by current user")

var ErrAccountManagementAuthenticatorNotOwnedbyToUser = apierrors.Invalid.WithReason("AccountManagementAuthenticatorNotOwnedByUser").New("authenticator not owned by current user")
var ErrAccountManagementSecondaryAuthenticatorIsRequired = apierrors.Invalid.WithReason("AccountManagementSecondaryAuthenticatorIsRequired").New("at least one secondary authenticator is needed")

func NewErrAccountManagementDuplicatedIdentity(originalErr error) error {
	return apierrors.AlreadyExists.WithReason("AccountManagementDuplicatedIdentity").Wrap(originalErr, "identity already exists")
}
