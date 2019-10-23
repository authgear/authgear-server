package forgotpwd

import (
	skyerr "github.com/skygeario/skygear-server/pkg/core/xskyerr"
)

var PasswordResetFailed = skyerr.Invalid.WithReason("PasswordResetFailed")

type resetFailCause string

const (
	InvalidCode        resetFailCause = "InvalidCode"
	UsedCode           resetFailCause = "UsedCode"
	ExpiredCode        resetFailCause = "ExpiredCode"
	PasswordNotMatched resetFailCause = "PasswordNotMatched"
)

func NewPasswordResetFailed(cause resetFailCause, msg string) error {
	return PasswordResetFailed.NewWithDetails(msg, skyerr.Details{"cause": skyerr.APIErrorString(cause)})
}
