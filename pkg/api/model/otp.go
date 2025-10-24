package model

type OTPDeliveryStatus string

const (
	OTPDeliveryStatusPending = "pending"
	OTPDeliveryStatusSending = "sending"
	OTPDeliveryStatusFailed  = "failed"
	OTPDeliveryStatusSent    = "sent"
)
