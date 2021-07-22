package config

var _ = FeatureConfigSchema.Add("OAuthFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"client": { "$ref": "#/$defs/OAuthClientFeatureConfig" }
	}
}
`)

type OAuthFeatureConfig struct {
	Client *OAuthClientFeatureConfig `json:"client,omitempty"`
}

var _ = FeatureConfigSchema.Add("OAuthClientFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"maximum": { "type": "integer" }
	}
}
`)

type OAuthClientFeatureConfig struct {
	Maximum *int `json:"maximum,omitempty"`
}

func (c *OAuthClientFeatureConfig) SetDefaults() {
	if c.Maximum == nil {
		c.Maximum = newInt(99)
	}
}
