package otp

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

var InvalidOTPCode = apierrors.Forbidden.WithReason("InvalidOTPCode")

var ErrCodeNotFound = InvalidOTPCode.NewWithCause("otp code is expired or invalid", apierrors.StringCause("CodeNotFound"))
var ErrInvalidCode = InvalidOTPCode.NewWithCause("invalid otp code", apierrors.StringCause("InvalidOTPCode"))
var ErrInvalidLoginLink = InvalidOTPCode.NewWithCause("invalid login link", apierrors.StringCause("InvalidLoginLink"))
var ErrInputRequired = InvalidOTPCode.NewWithCause("input not yet received", apierrors.StringCause("InputRequired"))

// FIXME: backward compat; should not use RateLimited
var ErrTooManyAttempts = ratelimit.RateLimited.NewWithInfo("too many verify OTP attempts", apierrors.Details{
	"bucket_name": ratelimit.TrackFailedOTPAttemptBucketName,
})
