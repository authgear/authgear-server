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

var _ MergeableFeatureConfig = &IdentityFeatureConfig{}

func (c *IdentityFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.Identity == nil {
		return c
	}

	merged := c
	if merged == nil {
		merged = &IdentityFeatureConfig{}
	}

	if layer.Identity.LoginID != nil {
		merged.LoginID = layer.Identity.LoginID
	}
	merged.OAuth = merged.OAuth.Merge(layer.Identity.OAuth)
	if layer.Identity.Biometric != nil {
		merged.Biometric = layer.Identity.Biometric
	}

	return merged
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
	// No omitempty: see the comment on CustomDomainFeatureConfig.Disabled
	// (feature_custom_domain.go).
	Disabled bool `json:"disabled"`
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
		c.MaximumProviders = new(99)
	}
}

func (c *OAuthSSOFeatureConfig) Merge(layer *OAuthSSOFeatureConfig) *OAuthSSOFeatureConfig {
	if c == nil && layer == nil {
		return nil
	}
	if c == nil {
		return layer
	}
	if layer == nil {
		return c
	}
	if layer.MaximumProviders != nil {
		c.MaximumProviders = layer.MaximumProviders
	}
	c.Providers = c.Providers.Merge(layer.Providers)
	return c
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

func (c *OAuthSSOProvidersFeatureConfig) Merge(layer *OAuthSSOProvidersFeatureConfig) *OAuthSSOProvidersFeatureConfig {
	if c == nil && layer == nil {
		return nil
	}
	if c == nil {
		return layer
	}
	if layer == nil {
		return c
	}
	if layer.Google != nil {
		c.Google = layer.Google
	}
	if layer.Facebook != nil {
		c.Facebook = layer.Facebook
	}
	if layer.Github != nil {
		c.Github = layer.Github
	}
	if layer.LinkedIn != nil {
		c.LinkedIn = layer.LinkedIn
	}
	if layer.Azureadv2 != nil {
		c.Azureadv2 = layer.Azureadv2
	}
	if layer.Azureadb2c != nil {
		c.Azureadb2c = layer.Azureadb2c
	}
	if layer.ADFS != nil {
		c.ADFS = layer.ADFS
	}
	if layer.Apple != nil {
		c.Apple = layer.Apple
	}
	if layer.Wechat != nil {
		c.Wechat = layer.Wechat
	}
	return c
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
	// No omitempty: see the comment on CustomDomainFeatureConfig.Disabled
	// (feature_custom_domain.go).
	Disabled bool `json:"disabled"`
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
		c.Disabled = new(false)
	}
}
