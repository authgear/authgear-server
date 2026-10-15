package config

var _ = FeatureConfigSchema.Add("PortalFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"auth_flow_builder_enabled": { "type": "boolean" }
	}
}
`)

// PortalFeatureConfig holds flags that only show or hide portal screens.
// Nothing outside the portal reads them; a flag that gates a capability
// belongs with that feature instead (e.g. fraud_protection.is_modifiable).
type PortalFeatureConfig struct {
	// AuthFlowBuilderEnabled gates the portal's auth flow builder while it is in testing.
	AuthFlowBuilderEnabled *bool `json:"auth_flow_builder_enabled,omitempty"`
}

var _ MergeableFeatureConfig = &PortalFeatureConfig{}

func (c *PortalFeatureConfig) SetDefaults() {
	if c.AuthFlowBuilderEnabled == nil {
		c.AuthFlowBuilderEnabled = new(false)
	}
}

func (c *PortalFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.Portal == nil {
		return c
	}

	merged := c
	if merged == nil {
		merged = &PortalFeatureConfig{}
	}

	if layer.Portal.AuthFlowBuilderEnabled != nil {
		merged.AuthFlowBuilderEnabled = layer.Portal.AuthFlowBuilderEnabled
	}

	return merged
}
