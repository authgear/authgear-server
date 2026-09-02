package otp

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
)

// OTPDeliveryStatusInternal is the delivery status recorded by the send path.
// The states are split by who can advance them, because that is what decides
// whether a reader may wait. No reader should look at the channel to decide it:
// whether a provider confirms asynchronously is recorded by the send path.
type OTPDeliveryStatusInternal string

const (
	// OTPDeliveryStatusInternalWaitingToSend means no delivery attempt has been
	// recorded yet. The send path records an attempt before the request that
	// created the code returns.
	OTPDeliveryStatusInternalWaitingToSend OTPDeliveryStatusInternal = "waiting_to_send"
	// OTPDeliveryStatusInternalWontSend means no message will ever be sent, and
	// that this is intended.
	OTPDeliveryStatusInternalWontSend OTPDeliveryStatusInternal = "wont_send"
	// OTPDeliveryStatusInternalWaitingForConfirmation means the message was handed
	// off to a provider that reports the outcome asynchronously.
	OTPDeliveryStatusInternalWaitingForConfirmation OTPDeliveryStatusInternal = "waiting_for_confirmation"
	// OTPDeliveryStatusInternalSent means the message was handed off and no further
	// outcome is expected.
	OTPDeliveryStatusInternalSent OTPDeliveryStatusInternal = "sent"
	// OTPDeliveryStatusInternalFailed is terminal for the code: a resend generates a
	// new code rather than clearing it.
	OTPDeliveryStatusInternalFailed OTPDeliveryStatusInternal = "failed"
)

func (s OTPDeliveryStatusInternal) ToAPIStatus() model.OTPDeliveryStatus {
	switch s {
	case OTPDeliveryStatusInternalWaitingToSend:
		// The request that created the code may still be recording its attempt, for
		// example a concurrent one in another tab.
		return model.OTPDeliveryStatusSending
	case OTPDeliveryStatusInternalWaitingForConfirmation:
		return model.OTPDeliveryStatusSending
	case OTPDeliveryStatusInternalWontSend:
		// We do not tell the end user that no message was sent, in the same way
		// InspectState pretends a code that does not exist was sent.
		return model.OTPDeliveryStatusSent
	case OTPDeliveryStatusInternalSent:
		return model.OTPDeliveryStatusSent
	case OTPDeliveryStatusInternalFailed:
		return model.OTPDeliveryStatusFailed
	default:
		panic(fmt.Errorf("unknown otp delivery status: %s", s))
	}
}
