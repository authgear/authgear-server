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

type UsageLimitPeriod string

const (
	UsageLimitPeriodDay   UsageLimitPeriod = "day"
	UsageLimitPeriodMonth UsageLimitPeriod = "month"
)

type UsageLimitConfig struct {
	Enabled *bool            `json:"enabled,omitempty"`
	Period  UsageLimitPeriod `json:"period,omitempty"`
	Quota   *int             `json:"quota,omitempty"`
}

func (c *UsageLimitConfig) IsEnabled() bool {
	if c == nil {
		return false
	}
	if c.Enabled == nil {
		return false
	}
	return *c.Enabled
}

func (c *UsageLimitConfig) GetQuota() int {
	if c == nil {
		return 0
	}
	if c.Quota == nil {
		return 0
	}
	return *c.Quota
}
