package config

var _ = FeatureConfigSchema.Add("AdminAPIFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"create_session_enabled": { "type": "boolean" }
	}
}
`)

type AdminAPIFeatureConfig struct {
	CreateSessionEnabled *bool `json:"create_session_enabled,omitempty"`
}

func (c *AdminAPIFeatureConfig) SetDefaults() {
	if c.CreateSessionEnabled == nil {
		c.CreateSessionEnabled = newBool(false)
	}
}
