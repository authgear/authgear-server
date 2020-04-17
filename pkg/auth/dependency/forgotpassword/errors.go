package forgotpassword

import (
	"errors"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

var ErrLoginIDNotFound = errors.New("login ID not found")
var ErrUnsupportedLoginIDType = errors.New("unsupported login ID type")

var PasswordResetFailed = skyerr.Invalid.WithReason("PasswordResetFailed")

var ErrInvalidCode = PasswordResetFailed.NewWithCause("invalid code", skyerr.StringCause("InvalidCode"))
var ErrUsedCode = PasswordResetFailed.NewWithCause("used code", skyerr.StringCause("UsedCode"))
var ErrExpiredCode = PasswordResetFailed.NewWithCause("expired code", skyerr.StringCause("ExpiredCode"))
