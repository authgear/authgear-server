package botprotection

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrVerificationFailed = apierrors.Forbidden.WithReason("ErrBotProtectionVerificationFailed").New("bot protection verification failed")

var ErrVerificationServiceUnavailable = apierrors.ServiceUnavailable.WithReason("ErrBotProtectionVerificationServiceUnavailable").New("bot protection service unavailable")
