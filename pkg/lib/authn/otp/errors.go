package otp

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var InvalidOTPCode = apierrors.Forbidden.WithReason("InvalidOTPCode")

var ErrCodeNotFound = InvalidOTPCode.NewWithCause("otp code is expired or invalid", apierrors.StringCause("CodeNotFound"))
var ErrInvalidCode = InvalidOTPCode.NewWithCause("invalid otp code", apierrors.StringCause("InvalidOTPCode"))
var ErrInvalidMagicLink = InvalidOTPCode.NewWithCause("invalid magic link", apierrors.StringCause("InvalidMagicLink"))
var ErrInputRequired = InvalidOTPCode.NewWithCause("input not yet received", apierrors.StringCause("InputRequired"))
