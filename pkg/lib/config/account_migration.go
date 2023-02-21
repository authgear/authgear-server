package config

var _ = Schema.Add("AccountMigrationConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"hook": { "$ref": "#/$defs/AccountMigrationHookConfig" }
	}
}
`)

type AccountMigrationConfig struct {
	Hook *AccountMigrationHookConfig `json:"hook,omitempty"`
}

var _ = Schema.Add("AccountMigrationHookConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"url": { "type": "string", "format": "x_hook_uri" },
		"timeout": { "type": "integer" }
	}
}
`)

type AccountMigrationHookConfig struct {
	URL     string          `json:"url,omitempty"`
	Timeout DurationSeconds `json:"timeout,omitempty"`
}

func (c *AccountMigrationHookConfig) SetDefaults() {
	if c.Timeout == 0 {
		c.Timeout = DurationSeconds(5)
	}
}
