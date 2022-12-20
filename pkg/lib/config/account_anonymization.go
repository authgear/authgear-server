package config

var _ = Schema.Add("AccountAnonymizationConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"grace_period_days": {
			"$ref": "#/$defs/DurationDays",
			"minimum": 1,
			"maximum": 180
		}
	}
}
`)

type AccountAnonymizationConfig struct {
	GracePeriod DurationDays `json:"grace_period_days,omitempty"`
}

func (c *AccountAnonymizationConfig) SetDefaults() {
	if c.GracePeriod == 0 {
		c.GracePeriod = DurationDays(30)
	}
}
