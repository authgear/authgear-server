package config

import "math"

var _ = Schema.Add("RateLimitConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"period": { "$ref": "#/$defs/DurationString" },
		"burst": { "type": "integer", "minimum": 1 }
	},
	"if": { "properties": { "enabled": { "const": true } }, "required": ["enabled"] },
	"then": { "required": ["period"] }
}
`)

var _ = FeatureConfigSchema.Add("RateLimitConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"period": { "$ref": "#/$defs/DurationString" },
		"burst": { "type": "integer", "minimum": 1 }
	},
	"if": { "properties": { "enabled": { "const": true } }, "required": ["enabled"] },
	"then": { "required": ["period"] }
}
`)

type RateLimitConfig struct {
	Enabled *bool          `json:"enabled,omitempty"`
	Period  DurationString `json:"period,omitempty"`
	Burst   int            `json:"burst,omitempty"`
}

func (c *RateLimitConfig) SetDefaults() {
	if c.Enabled != nil && *c.Enabled && c.Burst == 0 {
		c.Burst = 1
	}
}

func (c *RateLimitConfig) Rate() float64 {
	if c.Enabled == nil || !*c.Enabled {
		// Infinite rate if disabled.
		return math.Inf(1)
	}
	// request/seconds
	return float64(c.Burst) / c.Period.Duration().Seconds()
}

func (c *RateLimitConfig) IsEnabled() bool {
	if c.Enabled == nil || !*c.Enabled {
		return false
	}
	return true
}
