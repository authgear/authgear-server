package otp

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

var InvalidOTPCode = apierrors.Forbidden.WithReason("InvalidOTPCode")

var ErrCodeNotFound = InvalidOTPCode.NewWithCause("otp code is expired or invalid", apierrors.StringCause("CodeNotFound"))
var ErrInvalidCode = InvalidOTPCode.NewWithCause("invalid otp code", apierrors.StringCause("InvalidCode"))
var ErrConsumedCode = InvalidOTPCode.NewWithCause("used otp code", apierrors.StringCause("UsedCode"))

// FIXME: backward compat; should not use RateLimited
var ErrTooManyAttempts = ratelimit.RateLimited.NewWithInfo("too many verify OTP attempts", apierrors.Details{
	"bucket_name": "TrackFailedOTPAttemptBucket",
})

var ErrOTPDeliveryTimeout = apierrors.BadRequest.
	WithReason("OTPDeliveryTimeout").
	New("message status update not received within configured timeout")
