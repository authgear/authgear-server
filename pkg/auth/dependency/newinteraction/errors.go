package newinteraction

import "github.com/authgear/authgear-server/pkg/lib/api/apierrors"

var (
	ConfigurationViolated  = apierrors.Forbidden.WithReason("ConfigurationViolated")
	InvalidCredentials     = apierrors.Unauthorized.WithReason("InvalidCredentials")
	DuplicatedIdentity     = apierrors.AlreadyExists.WithReason("DuplicatedIdentity")
	InvalidIdentityRequest = apierrors.Invalid.WithReason("InvalidIdentityRequest")
)

var ErrInvalidCredentials = InvalidCredentials.New("invalid credentials")
var ErrDuplicatedIdentity = DuplicatedIdentity.New("identity already exists")
var ErrOAuthProviderNotFound = apierrors.NotFound.WithReason("OAuthProviderNotFound").New("oauth provider not found")
var ErrCannotRemoveLastIdentity = InvalidIdentityRequest.NewWithCause("cannot remove last identity", apierrors.StringCause("IdentityRequired"))
