package captcha

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var CaptchaFailed = apierrors.Invalid.WithReason("CaptchaFailed")

var ErrVerificationFailed = CaptchaFailed.NewWithCause("verification failed", apierrors.StringCause("VerificationFailed"))
