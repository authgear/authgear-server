package password

import (
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

var LoginIDAlreadyUsed = skyerr.AlreadyExists.WithReason("LoginIDAlreadyUsed")
var LoginIDNotFound = skyerr.NotFound.WithReason("LoginIDNotFound")
var InvalidCredentials = skyerr.Unauthorized.WithReason("InvalidCredentials")

var ErrLoginIDAlreadyUsed = LoginIDAlreadyUsed.New("login ID is already used")
var ErrLoginIDNotFound = LoginIDNotFound.New("login ID does not exist")
var ErrInvalidCredentials = InvalidCredentials.New("invalid credentials")
