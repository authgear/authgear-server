package config

var _ = Schema.Add("AuthenticationFlowRateLimitsConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"per_ip": { "$ref": "#/$defs/RateLimitConfig" }
	}
}
`)

type AuthenticationFlowRateLimitsConfig struct {
	PerIP *RateLimitConfig `json:"per_ip,omitempty"`
}

func (c *AuthenticationFlowRateLimitsConfig) SetDefaults() {
	if c.PerIP.Enabled == nil {
		c.PerIP = &RateLimitConfig{
			Enabled: newBool(true),
			Period:  "1m",
			Burst:   1200,
		}
	}
}
