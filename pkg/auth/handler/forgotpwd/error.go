package forgotpwd

import (
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
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
	return PasswordResetFailed.NewWithInfo(msg, skyerr.Details{"cause": cause})
}
