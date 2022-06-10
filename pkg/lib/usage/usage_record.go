package usage

// nolint:golint
type UsageRecordName string

const (
	ActiveUser          UsageRecordName = "active-user"
	SMSSent             UsageRecordName = "sms-sent"
	EmailSent           UsageRecordName = "email-sent"
	WhatsappOTPVerified UsageRecordName = "whatsapp-otp-verified"
)
