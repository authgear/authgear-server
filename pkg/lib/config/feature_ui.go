package config

var _ = FeatureConfigSchema.Add("UIFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"white_labeling": { "$ref": "#/$defs/WhiteLabelingFeatureConfig" },
		"allow_showing_logo_uri_as_project_logo": { "type": "boolean" }
	}
}
`)

type UIFeatureConfig struct {
	WhiteLabeling *WhiteLabelingFeatureConfig `json:"white_labeling,omitempty"`

	// AllowShowingLogoURIAsProjectLogo indicates whether the project allows showing logo URI as project logo.
	AllowShowingLogoURIAsProjectLogo *bool `json:"allow_showing_logo_uri_as_project_logo,omitempty"`
}

var _ MergeableFeatureConfig = &UIFeatureConfig{}

func (c *UIFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.UI == nil {
		return c
	}

	return layer.UI
}

func (c *UIFeatureConfig) SetDefaults() {
	if c.AllowShowingLogoURIAsProjectLogo == nil {
		c.AllowShowingLogoURIAsProjectLogo = newBool(false)
	}
}

func (c *UIFeatureConfig) GetAllowShowingLogoURIAsProjectLogo() bool {
	if c == nil {
		return false
	}

	if c.AllowShowingLogoURIAsProjectLogo == nil {
		return false
	}

	return *c.AllowShowingLogoURIAsProjectLogo
}

var _ = FeatureConfigSchema.Add("WhiteLabelingFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"disabled": { "type": "boolean" }
	}
}
`)

type WhiteLabelingFeatureConfig struct {
	Disabled bool `json:"disabled,omitempty"`
}
