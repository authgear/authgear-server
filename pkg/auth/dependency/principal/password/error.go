package password

import (
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

var LoginIDAlreadyUsed = skyerr.AlreadyExists.WithReason("LoginIDAlreadyUsed")
var InvalidCredentials = skyerr.Unauthorized.WithReason("InvalidCredentials")

var ErrLoginIDAlreadyUsed = LoginIDAlreadyUsed.New("login ID is used by another user")
var ErrInvalidCredentials = InvalidCredentials.New("invalid credentials")
