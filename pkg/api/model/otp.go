package model

type OTPDeliveryStatus string

const (
	OTPDeliveryStatusSending OTPDeliveryStatus = "sending"
	OTPDeliveryStatusFailed  OTPDeliveryStatus = "failed"
	OTPDeliveryStatusSent    OTPDeliveryStatus = "sent"
)
