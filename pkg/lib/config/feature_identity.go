package config

import (
	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	liboauthrelyingparty "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty"
)

var _ = FeatureConfigSchema.Add("IdentityFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"login_id": { "$ref": "#/$defs/LoginIDFeatureConfig" },
		"oauth": { "$ref": "#/$defs/OAuthSSOFeatureConfig" },
		"biometric": { "$ref": "#/$defs/BiometricFeatureConfig" }
	}
}
`)

type IdentityFeatureConfig struct {
	LoginID   *LoginIDFeatureConfig   `json:"login_id,omitempty"`
	OAuth     *OAuthSSOFeatureConfig  `json:"oauth,omitempty"`
	Biometric *BiometricFeatureConfig `json:"biometric,omitempty"`
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
		"github": { "$ref": "#/$defs/OAuthSSOProviderFeatureConfig" },
		"linkedin": { "$ref": "#/$defs/OAuthSSOProviderFeatureConfig" },
		"azureadv2": { "$ref": "#/$defs/OAuthSSOProviderFeatureConfig" },
		"azureadb2c": { "$ref": "#/$defs/OAuthSSOProviderFeatureConfig" },
		"adfs": { "$ref": "#/$defs/OAuthSSOProviderFeatureConfig" },
		"apple": { "$ref": "#/$defs/OAuthSSOProviderFeatureConfig" },
		"wechat": { "$ref": "#/$defs/OAuthSSOProviderFeatureConfig" }
	}
}
`)

type OAuthSSOProvidersFeatureConfig struct {
	Google     *OAuthSSOProviderFeatureConfig `json:"google,omitempty"`
	Facebook   *OAuthSSOProviderFeatureConfig `json:"facebook,omitempty"`
	Github     *OAuthSSOProviderFeatureConfig `json:"github,omitempty"`
	LinkedIn   *OAuthSSOProviderFeatureConfig `json:"linkedin,omitempty"`
	Azureadv2  *OAuthSSOProviderFeatureConfig `json:"azureadv2,omitempty"`
	Azureadb2c *OAuthSSOProviderFeatureConfig `json:"azureadb2c,omitempty"`
	ADFS       *OAuthSSOProviderFeatureConfig `json:"adfs,omitempty"`
	Apple      *OAuthSSOProviderFeatureConfig `json:"apple,omitempty"`
	Wechat     *OAuthSSOProviderFeatureConfig `json:"wechat,omitempty"`
}

func (c *OAuthSSOProvidersFeatureConfig) IsDisabled(cfg oauthrelyingparty.ProviderConfig) bool {
	switch cfg.Type() {
	case liboauthrelyingparty.TypeGoogle:
		return c.Google.Disabled
	case liboauthrelyingparty.TypeFacebook:
		return c.Facebook.Disabled
	case liboauthrelyingparty.TypeGithub:
		return c.Github.Disabled
	case liboauthrelyingparty.TypeLinkedin:
		return c.LinkedIn.Disabled
	case liboauthrelyingparty.TypeAzureADv2:
		return c.Azureadv2.Disabled
	case liboauthrelyingparty.TypeAzureADB2C:
		return c.Azureadb2c.Disabled
	case liboauthrelyingparty.TypeADFS:
		return c.ADFS.Disabled
	case liboauthrelyingparty.TypeApple:
		return c.Apple.Disabled
	case liboauthrelyingparty.TypeWechat:
		return c.Wechat.Disabled
	default:
		// Not a provider we recognize here. Allow it.
		return false
	}
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

var _ = FeatureConfigSchema.Add("BiometricFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"disabled": { "type": "boolean" }
	}
}
`)

type BiometricFeatureConfig struct {
	Disabled *bool `json:"disabled,omitempty"`
}

func (c *BiometricFeatureConfig) SetDefaults() {
	if c.Disabled == nil {
		c.Disabled = newBool(false)
	}
}
