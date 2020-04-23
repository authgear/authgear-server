package forgotpassword

import (
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

var PasswordResetFailed = skyerr.Invalid.WithReason("PasswordResetFailed")

var ErrInvalidCode = PasswordResetFailed.NewWithCause("invalid code", skyerr.StringCause("InvalidCode"))
var ErrUsedCode = PasswordResetFailed.NewWithCause("used code", skyerr.StringCause("UsedCode"))
var ErrExpiredCode = PasswordResetFailed.NewWithCause("expired code", skyerr.StringCause("ExpiredCode"))
