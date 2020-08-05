package verification

import (
	"github.com/authgear/authgear-server/pkg/core/skyerr"
)

var InvalidVerificationCode = skyerr.Forbidden.WithReason("InvalidVerificationCode")

var ErrCodeNotFound = InvalidVerificationCode.NewWithCause("verification code is expired or invalid", skyerr.StringCause("CodeNotFound"))
var ErrInvalidVerificationCode = InvalidVerificationCode.NewWithCause("invalid verification code", skyerr.StringCause("InvalidVerificationCode"))
