package accountmanagement

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrOAuthTokenInvalid = apierrors.Invalid.WithReason("AccountManagementOAuthTokenInvalid").New("invalid token")
var ErrOAuthStateNotBoundToToken = apierrors.Invalid.WithReason("AccountManagementOAuthStateNotBoundToToken").New("the state parameter in query is not bound to token")
var ErrOAuthTokenNotBoundToUser = apierrors.Invalid.WithReason("AccountManagementOAuthTokenNotBoundToUser").New("token is not bound to the current user")
