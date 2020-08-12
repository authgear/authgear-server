package verification

import (
	"github.com/authgear/authgear-server/pkg/lib/api/apierrors"
)

var InvalidVerificationCode = apierrors.Forbidden.WithReason("InvalidVerificationCode")

var ErrCodeNotFound = InvalidVerificationCode.NewWithCause("verification code is expired or invalid", apierrors.StringCause("CodeNotFound"))
var ErrInvalidVerificationCode = InvalidVerificationCode.NewWithCause("invalid verification code", apierrors.StringCause("InvalidVerificationCode"))
