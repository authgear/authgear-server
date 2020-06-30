package interaction

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/core/skyerr"
)

var ErrInteractionNotFound = errors.New("interaction not found")

var ErrInvalidStep = errors.New("step is invalid for current interaction state")

var ErrInvalidAction = errors.New("action is invalid for current interaction state")

var InvalidCredentials = skyerr.Unauthorized.WithReason("InvalidCredentials")

var ErrInvalidCredentials = InvalidCredentials.New("invalid credentials")

var DuplicatedIdentity = skyerr.AlreadyExists.WithReason("DuplicatedIdentity")

var ErrDuplicatedIdentity = DuplicatedIdentity.New("duplicate identity exists")

var IdentityNotFound = skyerr.NotFound.WithReason("IdentityNotFound")

var ErrIdentityNotFound = IdentityNotFound.New("identity not found")

var AuthenticatorNotFound = skyerr.NotFound.WithReason("AuthenticatorNotFound")

var ErrAuthenticatorNotFound = AuthenticatorNotFound.New("authenticator not found")

var ErrOOBOTPCooldown = skyerr.TooManyRequest.WithReason("OOBOTPCooldown").New("OOB OTP cooldown")

var InvalidIdentityRequest = skyerr.Invalid.WithReason("InvalidIdentityRequest")

var ErrCannotRemoveLastIdentity = InvalidIdentityRequest.NewWithCause("cannot remove last identity", skyerr.StringCause("IdentityRequired"))
