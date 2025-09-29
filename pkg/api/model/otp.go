package model

type OTPDeliveryStatus string

const (
	OTPDeliveryStatusSending = "sending"
	OTPDeliveryStatusFailed  = "failed"
	OTPDeliveryStatusSent    = "sent"
)
