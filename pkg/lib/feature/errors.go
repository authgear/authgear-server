package feature

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrFeatureDisabledSendingSMS = apierrors.Forbidden.WithReason("SMSNotSupported").New("sending sms is not supported")
