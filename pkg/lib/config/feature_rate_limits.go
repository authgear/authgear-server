package config

var _ = FeatureConfigSchema.Add("RateLimitsFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"disabled": { "type": "boolean" },
		"sms": { "$ref": "#/$defs/RateLimitConfig" },
		"sms_per_ip": { "$ref": "#/$defs/RateLimitConfig" },
		"sms_per_target": { "$ref": "#/$defs/RateLimitConfig" },
		"email": { "$ref": "#/$defs/RateLimitConfig" },
		"email_per_ip": { "$ref": "#/$defs/RateLimitConfig" },
		"email_per_target": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

type RateLimitsFeatureConfig struct {
	Disabled       bool             `json:"disabled,omitempty"`
	SMS            *RateLimitConfig `json:"sms,omitempty"`
	SMSPerIP       *RateLimitConfig `json:"sms_per_ip,omitempty"`
	SMSPerTarget   *RateLimitConfig `json:"sms_per_target,omitempty"`
	Email          *RateLimitConfig `json:"email,omitempty"`
	EmailPerIP     *RateLimitConfig `json:"email_per_ip,omitempty"`
	EmailPerTarget *RateLimitConfig `json:"email_per_target,omitempty"`
}

func (c *RateLimitsFeatureConfig) SetDefaults() {
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
