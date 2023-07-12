package config

var _ = Schema.Add("OTPLegacyConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"ratelimit": { "$ref": "#/$defs/OTPLegacyRatelimitConfig" }
	}
}
`)

type OTPLegacyConfig struct {
	Ratelimit *OTPLegacyRatelimitConfig `json:"ratelimit,omitempty"`
}

var _ = Schema.Add("OTPLegacyRatelimitConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"failed_attempt": { "$ref": "#/$defs/OTPLegacyFailedAttemptConfig" }
	}
}
`)

type OTPLegacyRatelimitConfig struct {
	FailedAttempt *OTPLegacyFailedAttemptConfig `json:"failed_attempt,omitempty"`
}

var _ = Schema.Add("OTPLegacyFailedAttemptConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"size": {
			"type": "integer",
			"minimum": 1,
			"maximum": 10
		},
		"reset_period": { "$ref": "#/$defs/DurationString" }
	}
}
`)

type OTPLegacyFailedAttemptConfig struct {
	Enabled     bool           `json:"enabled,omitempty"`
	Size        int            `json:"size,omitempty"`
	ResetPeriod DurationString `json:"reset_period,omitempty"`
}

func (c *OTPLegacyFailedAttemptConfig) SetDefaults() {
	if c.Enabled {
		if c.Size == 0 {
			c.Size = 5
		}
		if c.ResetPeriod == "" {
			c.ResetPeriod = "20m"
		}
	}
}
