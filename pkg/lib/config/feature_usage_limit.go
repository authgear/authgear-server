package config

var _ = FeatureConfigSchema.Add("UsageLimitConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"period": { "$ref": "#/$defs/UsageLimitPeriod" },
		"quota": { "type": "integer", "minimum": 0 }
	},
	"if": { "properties": { "enabled": { "const": true } }, "required": ["enabled"] },
	"then": { "required": ["period", "quota"] }
}
`)

var _ = FeatureConfigSchema.Add("UsageLimitPeriod", `
{
	"type": "string",
	"enum": ["day", "month"]
}
`)

type Deprecated_UsageLimitPeriod string

const (
	Deprecated_UsageLimitPeriodDay   Deprecated_UsageLimitPeriod = "day"
	Deprecated_UsageLimitPeriodMonth Deprecated_UsageLimitPeriod = "month"
)

type Deprecated_UsageLimitConfig struct {
	Enabled *bool                       `json:"enabled,omitempty"`
	Period  Deprecated_UsageLimitPeriod `json:"period,omitempty"`
	Quota   *int                        `json:"quota,omitempty"`
}

func (c *Deprecated_UsageLimitConfig) IsEnabled() bool {
	if c == nil {
		return false
	}
	if c.Enabled == nil {
		return false
	}
	return *c.Enabled
}

func (c *Deprecated_UsageLimitConfig) GetQuota() int {
	if c == nil {
		return 0
	}
	if c.Quota == nil {
		return 0
	}
	return *c.Quota
}
