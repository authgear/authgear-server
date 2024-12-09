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

var _ MergeableFeatureConfig = &OAuthFeatureConfig{}

func (c *OAuthFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.OAuth == nil {
		return c
	}
	return layer.OAuth
}

var _ = FeatureConfigSchema.Add("OAuthClientFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"maximum": { "type": "integer" },
		"soft_maximum": { "type": "integer" },
		"custom_ui_enabled": { "type": "boolean" },
		"app2app_enabled": { "type": "boolean" }
	}
}
`)

type OAuthClientFeatureConfig struct {
	Maximum         *int `json:"maximum,omitempty"`
	SoftMaximum     *int `json:"soft_maximum,omitempty"`
	CustomUIEnabled bool `json:"custom_ui_enabled,omitempty"`
	App2AppEnabled  bool `json:"app2app_enabled,omitempty"`
}

func (c *OAuthClientFeatureConfig) SetDefaults() {
	if c.Maximum == nil {
		c.Maximum = newInt(99)
	}

	if c.SoftMaximum == nil {
		c.SoftMaximum = newInt(99)
	}
}
