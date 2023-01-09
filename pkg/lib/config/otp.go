package config

var _ = Schema.Add("OTPConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"sms": { "$ref": "#/$defs/OTPSMSConfig" }
	}
}
`)

type OTPConfig struct {
	SMS *OTPSMSConfig `json:"sms,omitempty"`
}

var _ = Schema.Add("OTPSMSConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"resend_cooldown_seconds": {
			"type": "integer",
			"enum": [60, 120]
		}
	}
}
`)

type OTPSMSConfig struct {
	ResendCooldownSeconds DurationSeconds `json:"resend_cooldown_seconds,omitempty"`
}

func (c *OTPSMSConfig) SetDefaults() {
	if c.ResendCooldownSeconds == 0 {
		c.ResendCooldownSeconds = DurationSeconds(60)
	}
}
