package forgotpassword

import (
	"github.com/authgear/authgear-server/pkg/core/skyerr"
)

var PasswordResetFailed = skyerr.Invalid.WithReason("PasswordResetFailed")

var ErrInvalidCode = PasswordResetFailed.NewWithCause("invalid code", skyerr.StringCause("InvalidCode"))
var ErrUsedCode = PasswordResetFailed.NewWithCause("used code", skyerr.StringCause("UsedCode"))
var ErrExpiredCode = PasswordResetFailed.NewWithCause("expired code", skyerr.StringCause("ExpiredCode"))

var SendCodeFailed = skyerr.Invalid.WithReason("ForgotPasswordFailed")

var ErrUserNotFound = SendCodeFailed.NewWithCause("specified user not found", skyerr.StringCause("UserNotFound"))
var ErrNoPassword = SendCodeFailed.NewWithCause("specified user cannot be logged in using password", skyerr.StringCause("NoPassword"))
