package newinteraction

import "github.com/authgear/authgear-server/pkg/core/skyerr"

var (
	ConfigurationViolated = skyerr.Forbidden.WithReason("ConfigurationViolated")
	InvalidCredentials    = skyerr.Unauthorized.WithReason("InvalidCredentials")
	DuplicatedIdentity    = skyerr.AlreadyExists.WithReason("DuplicatedIdentity")
)

var ErrInvalidCredentials = InvalidCredentials.New("invalid credentials")
var ErrDuplicatedIdentity = DuplicatedIdentity.New("identity already exists")
var ErrOAuthProviderNotFound = skyerr.NotFound.WithReason("OAuthProviderNotFound").New("oauth provider not found")
