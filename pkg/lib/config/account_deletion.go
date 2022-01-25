package config

var _ = Schema.Add("AccountDeletionConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"scheduled_by_end_user_enabled": { "type": "boolean" },
		"grace_period_days": {
			"$ref": "#/$defs/DurationDays",
			"minimum": 1,
			"maximum": 180
		}
	}
}
`)

type AccountDeletionConfig struct {
	ScheduledByEndUserEnabled bool         `json:"scheduled_by_end_user_enabled,omitempty"`
	GracePeriod               DurationDays `json:"grace_period_days,omitempty"`
}

func (c *AccountDeletionConfig) SetDefaults() {
	if c.GracePeriod == 0 {
		c.GracePeriod = DurationDays(30)
	}
}
