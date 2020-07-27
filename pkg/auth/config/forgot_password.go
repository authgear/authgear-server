package config

var _ = Schema.Add("ForgotPasswordConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"email_message": { "$ref": "#/$defs/EmailMessageConfig" },
		"sms_message": { "$ref": "#/$defs/SMSMessageConfig" },
		"reset_code_expiry_seconds": { "$ref": "#/$defs/DurationSeconds" }
	}
}
`)

type ForgotPasswordConfig struct {
	Enabled         *bool              `json:"enabled,omitempty"`
	EmailMessage    EmailMessageConfig `json:"email_message,omitempty"`
	SMSMessage      SMSMessageConfig   `json:"sms_message,omitempty"`
	ResetCodeExpiry DurationSeconds    `json:"reset_code_expiry_seconds,omitempty"`
}

func (c *ForgotPasswordConfig) SetDefaults() {
	if c.Enabled == nil {
		c.Enabled = newBool(true)
	}

	if c.EmailMessage["subject"] == "" {
		c.EmailMessage["subject"] = "Reset password instruction"
	}
	if c.ResetCodeExpiry == 0 {
		// https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html#step-3-send-a-token-over-a-side-channel
		// OWASP suggests the lifetime is no more than 20 minutes
		c.ResetCodeExpiry = DurationSeconds(1200)
	}
}
