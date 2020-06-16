package config

var _ = Schema.Add("ForgotPasswordConfig", `
{
	"type": "object",
	"properties": {
		"email_message": { "$ref": "#/$defs/EmailMessageConfig" },
		"sms_message": { "$ref": "#/$defs/SMSMessageConfig" },
		"reset_code_expiry_seconds": { "$ref": "#/$defs/DurationSeconds" }
	}
}
`)

type ForgotPasswordConfig struct {
	EmailMessage    EmailMessageConfig `json:"email_message,omitempty"`
	SMSMessage      SMSMessageConfig   `json:"sms_message,omitempty"`
	ResetCodeExpiry DurationSeconds    `json:"reset_code_expiry_seconds,omitempty"`
}
