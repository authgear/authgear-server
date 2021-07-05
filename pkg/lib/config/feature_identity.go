package config

var _ = FeatureConfigSchema.Add("IdentityFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"login_id": { "$ref": "#/$defs/LoginIDFeatureConfig" }
	}
}
`)

type IdentityFeatureConfig struct {
	LoginID *LoginIDFeatureConfig `json:"login_id,omitempty"`
}

var _ = FeatureConfigSchema.Add("LoginIDFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"types": { "$ref": "#/$defs/LoginIDTypesFeatureConfig" }
	}
}
`)

type LoginIDFeatureConfig struct {
	Types *LoginIDTypesFeatureConfig `json:"types,omitempty"`
}

var _ = FeatureConfigSchema.Add("LoginIDTypesFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"phone": { "$ref": "#/$defs/LoginIDPhoneFeatureConfig" }
	}
}
`)

type LoginIDTypesFeatureConfig struct {
	Phone *LoginIDPhoneFeatureConfig `json:"phone,omitempty"`
}

var _ = FeatureConfigSchema.Add("LoginIDPhoneFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"disabled": { "type": "boolean" }
	}
}
`)

type LoginIDPhoneFeatureConfig struct {
	Disabled bool `json:"disabled,omitempty"`
}
