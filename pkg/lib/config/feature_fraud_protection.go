package config

var _ = FeatureConfigSchema.Add("FraudProtectionFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"is_modifiable": { "type": "boolean" }
	}
}
`)

type FraudProtectionFeatureConfig struct {
	IsModifiable *bool `json:"is_modifiable,omitempty"`
}

var _ MergeableFeatureConfig = &FraudProtectionFeatureConfig{}

func (c *FraudProtectionFeatureConfig) SetDefaults() {
	if c.IsModifiable == nil {
		c.IsModifiable = newBool(false)
	}
}

func (c *FraudProtectionFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.FraudProtection == nil {
		return c
	}
	return layer.FraudProtection
}
