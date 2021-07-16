package config

var _ = FeatureConfigSchema.Add("IdentityFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"login_id": { "$ref": "#/$defs/LoginIDFeatureConfig" },
		"oauth": { "$ref": "#/$defs/OAuthSSOFeatureConfig" }
	}
}
`)

type IdentityFeatureConfig struct {
	LoginID *LoginIDFeatureConfig  `json:"login_id,omitempty"`
	OAuth   *OAuthSSOFeatureConfig `json:"oauth,omitempty"`
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

var _ = FeatureConfigSchema.Add("OAuthSSOFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"maximum_providers": { "type": "integer" },
		"providers": { "$ref": "#/$defs/OAuthSSOProvidersFeatureConfig" }
	}
}
`)

type OAuthSSOFeatureConfig struct {
	MaximumProviders *int                            `json:"maximum_providers,omitempty"`
	Providers        *OAuthSSOProvidersFeatureConfig `json:"providers,omitempty"`
}

func (c *OAuthSSOFeatureConfig) SetDefaults() {
	if c.MaximumProviders == nil {
		c.MaximumProviders = newInt(99)
	}
}

var _ = FeatureConfigSchema.Add("OAuthSSOProvidersFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"google": { "$ref": "#/$defs/OAuthSSOProviderFeatureConfig" },
		"facebook": { "$ref": "#/$defs/OAuthSSOProviderFeatureConfig" },
		"linkedin": { "$ref": "#/$defs/OAuthSSOProviderFeatureConfig" },
		"azureadv2": { "$ref": "#/$defs/OAuthSSOProviderFeatureConfig" },
		"adfs": { "$ref": "#/$defs/OAuthSSOProviderFeatureConfig" },
		"apple": { "$ref": "#/$defs/OAuthSSOProviderFeatureConfig" },
		"wechat": { "$ref": "#/$defs/OAuthSSOProviderFeatureConfig" }
	}
}
`)

type OAuthSSOProvidersFeatureConfig struct {
	Google    *OAuthSSOProviderFeatureConfig `json:"google,omitempty"`
	Facebook  *OAuthSSOProviderFeatureConfig `json:"facebook,omitempty"`
	LinkedIn  *OAuthSSOProviderFeatureConfig `json:"linkedin,omitempty"`
	Azureadv2 *OAuthSSOProviderFeatureConfig `json:"azureadv2,omitempty"`
	ADFS      *OAuthSSOProviderFeatureConfig `json:"adfs,omitempty"`
	Apple     *OAuthSSOProviderFeatureConfig `json:"apple,omitempty"`
	Wechat    *OAuthSSOProviderFeatureConfig `json:"wechat,omitempty"`
}

var _ = FeatureConfigSchema.Add("OAuthSSOProviderFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"disabled": { "type": "boolean" }
	}
}
`)

type OAuthSSOProviderFeatureConfig struct {
	Disabled bool `json:"disabled,omitempty"`
}
