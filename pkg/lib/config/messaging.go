package config

var _ = Schema.Add("MessagingConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"sms_provider": { "$ref": "#/$defs/SMSProvider" },
		"whatsapp": { "$ref": "#/$defs/WhatsappConfig" },
		"rate_limits": { "$ref": "#/$defs/MessagingRateLimitsConfig" }
	}
}
`)

type MessagingConfig struct {
	SMSProvider SMSProvider                `json:"sms_provider,omitempty"`
	Whatsapp    *WhatsappConfig            `json:"whatsapp,omitempty"`
	RateLimits  *MessagingRateLimitsConfig `json:"rate_limits,omitempty"`
}

var _ = Schema.Add("SMSProvider", `
{
	"type": "string",
	"enum": ["nexmo", "twilio", "custom"]
}
`)

type SMSProvider string

const (
	SMSProviderNexmo  SMSProvider = "nexmo"
	SMSProviderTwilio SMSProvider = "twilio"
	SMSProviderCustom SMSProvider = "custom"
)

var _ = Schema.Add("MessagingRateLimitsConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"sms": { "$ref": "#/$defs/RateLimitConfig" },
		"sms_per_ip": { "$ref": "#/$defs/RateLimitConfig" },
		"sms_per_target": { "$ref": "#/$defs/RateLimitConfig" },
		"email": { "$ref": "#/$defs/RateLimitConfig" },
		"email_per_ip": { "$ref": "#/$defs/RateLimitConfig" },
		"email_per_target": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

type MessagingRateLimitsConfig struct {
	SMS            *RateLimitConfig `json:"sms,omitempty"`
	SMSPerIP       *RateLimitConfig `json:"sms_per_ip,omitempty"`
	SMSPerTarget   *RateLimitConfig `json:"sms_per_target,omitempty"`
	Email          *RateLimitConfig `json:"email,omitempty"`
	EmailPerIP     *RateLimitConfig `json:"email_per_ip,omitempty"`
	EmailPerTarget *RateLimitConfig `json:"email_per_target,omitempty"`
}

func (c *MessagingRateLimitsConfig) SetDefaults() {
	if c.SMSPerIP.Enabled == nil {
		c.SMSPerIP = &RateLimitConfig{
			Enabled: newBool(true),
			Period:  "1m",
			Burst:   60,
		}
	}
	if c.SMSPerTarget.Enabled == nil {
		c.SMSPerTarget = &RateLimitConfig{
			Enabled: newBool(true),
			Period:  "1h",
			Burst:   10,
		}
	}
	if c.EmailPerIP.Enabled == nil {
		c.EmailPerIP = &RateLimitConfig{
			Enabled: newBool(true),
			Period:  "1m",
			Burst:   200,
		}
	}
	if c.EmailPerTarget.Enabled == nil {
		c.EmailPerTarget = &RateLimitConfig{
			Enabled: newBool(true),
			Period:  "24h",
			Burst:   50,
		}
	}
}
