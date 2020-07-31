package verification

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/core/skyerr"
)

var InvalidVerificationCode = skyerr.Forbidden.WithReason("InvalidVerificationCode")

var ErrCodeNotFound = errors.New("verification code not found")
var ErrInvalidVerificationCode = InvalidVerificationCode.New("invalid verification code")
