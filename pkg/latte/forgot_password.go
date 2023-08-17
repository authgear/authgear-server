package latte

type ForgotPasswordChannel string

const (
	ForgotPasswordChannelEmail ForgotPasswordChannel = "email"
	ForgotPasswordChannelSMS   ForgotPasswordChannel = "sms"
)
