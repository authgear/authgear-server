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

	var merged *OAuthFeatureConfig = c
	if merged == nil {
		merged = &OAuthFeatureConfig{}
	}

	merged.Client = merged.Client.Merge(layer.OAuth.Client)

	return merged
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
	Maximum         *int  `json:"maximum,omitempty"`
	SoftMaximum     *int  `json:"soft_maximum,omitempty"`
	CustomUIEnabled *bool `json:"custom_ui_enabled,omitempty"`
	App2AppEnabled  *bool `json:"app2app_enabled,omitempty"`
}

func (c *OAuthClientFeatureConfig) SetDefaults() {
	if c.Maximum == nil {
		c.Maximum = newInt(99)
	}

	if c.SoftMaximum == nil {
		c.SoftMaximum = newInt(99)
	}

	if c.CustomUIEnabled == nil {
		c.CustomUIEnabled = newBool(false)
	}

	if c.App2AppEnabled == nil {
		c.App2AppEnabled = newBool(false)
	}
}

func (c *OAuthClientFeatureConfig) Merge(layer *OAuthClientFeatureConfig) *OAuthClientFeatureConfig {
	if c == nil && layer == nil {
		return nil
	}
	if c == nil {
		return layer
	}
	if layer == nil {
		return c
	}
	if layer.Maximum != nil {
		c.Maximum = layer.Maximum
	}
	if layer.SoftMaximum != nil {
		c.SoftMaximum = layer.SoftMaximum
	}
	if layer.App2AppEnabled != nil {
		c.App2AppEnabled = layer.App2AppEnabled
	}
	if layer.CustomUIEnabled != nil {
		c.CustomUIEnabled = layer.CustomUIEnabled
	}
	return c
}
