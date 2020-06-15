package config

type ForgotPasswordConfig struct {
	EmailMessage    EmailMessageConfig `json:"email_message,omitempty"`
	SMSMessage      SMSMessageConfig   `json:"sms_message,omitempty"`
	ResetCodeExpiry DurationSeconds    `json:"reset_code_expiry_seconds,omitempty"`
}
