package userverify

import (
	"errors"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

var ErrCodeNotFound = errors.New("user verification code not found")
var ErrUnknownLoginIDKey = errors.New("login ID key not configured")

var UserVerificationFailed = skyerr.Invalid.WithReason("UserVerificationFailed")

type verificationFailCause string

const (
	InvalidCode verificationFailCause = "InvalidCode"
	UsedCode    verificationFailCause = "UsedCode"
	ExpiredCode verificationFailCause = "ExpiredCode"
)

func NewUserVerificationFailed(cause verificationFailCause, msg string) error {
	return UserVerificationFailed.NewWithCause(msg, skyerr.StringCause(cause))
}
