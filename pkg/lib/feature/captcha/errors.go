package captcha

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var CaptchaFailed = apierrors.Invalid.WithReason("CaptchaFailed")

var ErrVerfificationFailed = CaptchaFailed.NewWithCause("verification failed", apierrors.StringCause("VerificationFailed"))
