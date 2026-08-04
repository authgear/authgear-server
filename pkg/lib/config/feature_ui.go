package config

import "github.com/authgear/authgear-server/pkg/util/phone"

var _ = FeatureConfigSchema.Add("UIFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"white_labeling": { "$ref": "#/$defs/WhiteLabelingFeatureConfig" },
		"phone_input": { "$ref": "#/$defs/PhoneInputFeatureConfig" }
	}
}
`)

type UIFeatureConfig struct {
	WhiteLabeling *WhiteLabelingFeatureConfig `json:"white_labeling,omitempty"`
	PhoneInput    *PhoneInputFeatureConfig    `json:"phone_input,omitempty"`
}

var _ MergeableFeatureConfig = &UIFeatureConfig{}

func (c *UIFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.UI == nil {
		return c
	}

	merged := c
	if merged == nil {
		merged = &UIFeatureConfig{}
	}

	if layer.UI.WhiteLabeling != nil {
		merged.WhiteLabeling = layer.UI.WhiteLabeling
	}
	if layer.UI.PhoneInput != nil {
		merged.PhoneInput = layer.UI.PhoneInput
	}

	return merged
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
	// No omitempty: see the comment on CustomDomainFeatureConfig.Disabled.
	Disabled bool `json:"disabled"`
}

var _ = FeatureConfigSchema.Add("PhoneInputFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"allowlist": { "type": "array", "items": { "$ref": "#/$defs/ISO31661Alpha2" } }
	}
}
`)

type PhoneInputFeatureConfig struct {
	// omitzero, not omitempty: nil (unset, inherit from plan) and an
	// explicit empty slice (allow all countries) are semantically distinct
	// here (see ApplyFeatureConfigConstraints/IntersectAllowlist). omitempty
	// treats both as "empty" and drops the field either way, which loses
	// the explicit-empty signal when this struct is re-marshaled during the
	// merge fold's final yaml.Marshal step (configsource/resources.go's
	// viewEffectiveResource) -- turning an explicit "allow all" override
	// back into an absent field, indistinguishable from "not set", once
	// re-parsed. omitzero only omits the true zero value (nil), preserving
	// AllowList: []string{} through that round trip.
	AllowList []string `json:"allowlist,omitzero"`
}

var _ = FeatureConfigSchema.Add("ISO31661Alpha2", phone.JSONSchemaString)
