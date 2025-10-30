package otp

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type OTPDeliveryStatusInternal string

const (
	OTPDeliveryStatusInternalPending OTPDeliveryStatusInternal = "pending"
	OTPDeliveryStatusInternalSending OTPDeliveryStatusInternal = "sending"
	OTPDeliveryStatusInternalFailed  OTPDeliveryStatusInternal = "failed"
	OTPDeliveryStatusInternalSent    OTPDeliveryStatusInternal = "sent"
)

func (s OTPDeliveryStatusInternal) ToAPIStatus() model.OTPDeliveryStatus {
	switch s {
	case OTPDeliveryStatusInternalPending:
		return model.OTPDeliveryStatusSending
	case OTPDeliveryStatusInternalSending:
		return model.OTPDeliveryStatusSending
	case OTPDeliveryStatusInternalFailed:
		return model.OTPDeliveryStatusFailed
	case OTPDeliveryStatusInternalSent:
		return model.OTPDeliveryStatusSent
	default:
		panic(fmt.Errorf("unknown otp delivery status: %s", s))
	}
}
