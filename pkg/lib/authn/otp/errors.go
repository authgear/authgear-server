package otp

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

var InvalidOTPCode = apierrors.Forbidden.WithReason("InvalidOTPCode")
var InvalidWhatsappUser = apierrors.BadRequest.WithReason("InvalidWhatsappUser")
var NoAvailableWhatsappClient = apierrors.BadRequest.WithReason("NoAvailableWhatsappClient")

var ErrCodeNotFound = InvalidOTPCode.NewWithCause("otp code is expired or invalid", apierrors.StringCause("CodeNotFound"))
var ErrInvalidCode = InvalidOTPCode.NewWithCause("invalid otp code", apierrors.StringCause("InvalidCode"))
var ErrConsumedCode = InvalidOTPCode.NewWithCause("used otp code", apierrors.StringCause("UsedCode"))

var ErrInvalidWhatsappUser = InvalidWhatsappUser.New("invalid whatsapp user")
var ErrNoAvailableWhatsappClient = NoAvailableWhatsappClient.New("no available whatsapp client")

// FIXME: backward compat; should not use RateLimited
var ErrTooManyAttempts = ratelimit.RateLimited.NewWithInfo("too many verify OTP attempts", apierrors.Details{
	"bucket_name": "TrackFailedOTPAttemptBucket",
})
