package botprotection

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrVerificationFailed = apierrors.Forbidden.WithReason("BotProtectionVerificationFailed").New("bot protection verification failed")

var ErrVerificationServiceUnavailable = apierrors.ServiceUnavailable.WithReason("BotProtectionVerificationServiceUnavailable").New("bot protection service unavailable")
