package config

import (
	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/api/model"
)

var _ = Schema.Add("IdentityConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"login_id": { "$ref": "#/$defs/LoginIDConfig" },
		"oauth": { "$ref": "#/$defs/OAuthSSOConfig" },
		"biometric": { "$ref": "#/$defs/BiometricConfig" },
		"on_conflict": { "$ref": "#/$defs/IdentityConflictConfig" }
	}
}
`)

type IdentityConfig struct {
	LoginID    *LoginIDConfig          `json:"login_id,omitempty"`
	OAuth      *OAuthSSOConfig         `json:"oauth,omitempty"`
	Biometric  *BiometricConfig        `json:"biometric,omitempty"`
	OnConflict *IdentityConflictConfig `json:"on_conflict,omitempty"`
}

var _ = Schema.Add("LoginIDConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"types": { "$ref": "#/$defs/LoginIDTypesConfig" },
		"keys": { "type": "array", "items": { "$ref": "#/$defs/LoginIDKeyConfig" } }
	}
}
`)

type LoginIDConfig struct {
	Types *LoginIDTypesConfig `json:"types,omitempty"`
	Keys  []LoginIDKeyConfig  `json:"keys,omitempty"`
}

func (c *LoginIDConfig) SetDefaults() {
	if c.Keys == nil {
		c.Keys = []LoginIDKeyConfig{
			{Type: model.LoginIDKeyTypeEmail},
		}
		for i := range c.Keys {
			c.Keys[i].SetDefaults()
		}
	}
}

func (c *LoginIDConfig) GetKeyConfig(key string) (*LoginIDKeyConfig, bool) {
	for _, config := range c.Keys {
		if config.Key == key {
			return &config, true
		}
	}

	return nil, false
}

var _ = Schema.Add("LoginIDTypesConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"email": { "$ref": "#/$defs/LoginIDEmailConfig" },
		"username": { "$ref": "#/$defs/LoginIDUsernameConfig" }
	}
}
`)

type LoginIDTypesConfig struct {
	Email    *LoginIDEmailConfig    `json:"email,omitempty"`
	Username *LoginIDUsernameConfig `json:"username,omitempty"`
}

var _ = Schema.Add("LoginIDEmailConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"case_sensitive": { "type": "boolean" },
		"block_plus_sign": { "type": "boolean" },
		"ignore_dot_sign": { "type": "boolean" },
		"domain_blocklist_enabled" : {"type": "boolean"},
		"domain_allowlist_enabled" : {"type": "boolean"},
		"block_free_email_provider_domains" : {"type": "boolean"}
	},
	"allOf": [
		{
			"if": {
				"properties": {
					"domain_blocklist_enabled": { "enum": [true] }
				},
				"required": ["domain_blocklist_enabled"]
			},
			"then": {
				"properties": {
					"domain_allowlist_enabled": { "enum": [false] }
				}
			}
		},
		{
			"if": {
				"properties": {
					"block_free_email_provider_domains": { "enum": [true] }
				},
				"required": ["block_free_email_provider_domains"]
			},
			"then": {
				"properties": {
					"domain_blocklist_enabled": { "enum": [true] }
				},
				"required": ["domain_blocklist_enabled"]
			}
		}
	]
}
`)

type LoginIDEmailConfig struct {
	CaseSensitive                 *bool `json:"case_sensitive,omitempty"`
	BlockPlusSign                 *bool `json:"block_plus_sign,omitempty"`
	IgnoreDotSign                 *bool `json:"ignore_dot_sign,omitempty"`
	DomainBlocklistEnabled        *bool `json:"domain_blocklist_enabled,omitempty"`
	DomainAllowlistEnabled        *bool `json:"domain_allowlist_enabled,omitempty"`
	BlockFreeEmailProviderDomains *bool `json:"block_free_email_provider_domains,omitempty"`
}

func (c *LoginIDEmailConfig) SetDefaults() {
	if c.CaseSensitive == nil {
		c.CaseSensitive = newBool(false)
	}
	if c.BlockPlusSign == nil {
		c.BlockPlusSign = newBool(false)
	}
	if c.IgnoreDotSign == nil {
		c.IgnoreDotSign = newBool(false)
	}
	if c.DomainBlocklistEnabled == nil {
		c.DomainBlocklistEnabled = newBool(false)
	}
	if c.DomainAllowlistEnabled == nil {
		c.DomainAllowlistEnabled = newBool(false)
	}
	if c.BlockFreeEmailProviderDomains == nil {
		c.BlockFreeEmailProviderDomains = newBool(false)
	}
}

var _ = Schema.Add("LoginIDUsernameConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"block_reserved_usernames": { "type": "boolean" },
		"exclude_keywords_enabled": { "type": "boolean" },
		"ascii_only": { "type": "boolean" },
		"case_sensitive": { "type": "boolean" }
	}
}
`)

type LoginIDUsernameConfig struct {
	BlockReservedUsernames *bool `json:"block_reserved_usernames,omitempty"`
	ExcludeKeywordsEnabled *bool `json:"exclude_keywords_enabled,omitempty"`
	ASCIIOnly              *bool `json:"ascii_only,omitempty"`
	CaseSensitive          *bool `json:"case_sensitive,omitempty"`
}

func (c *LoginIDUsernameConfig) SetDefaults() {
	if c.BlockReservedUsernames == nil {
		c.BlockReservedUsernames = newBool(true)
	}
	if c.ExcludeKeywordsEnabled == nil {
		c.ExcludeKeywordsEnabled = newBool(false)
	}
	if c.ASCIIOnly == nil {
		c.ASCIIOnly = newBool(true)
	}
	if c.CaseSensitive == nil {
		c.CaseSensitive = newBool(false)
	}
}

var _ = Schema.Add("LoginIDKeyConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"key": { "type": "string" },
		"type": { "$ref": "#/$defs/LoginIDKeyType" },
		"max_length": { "type": "integer" },
		"modify_disabled": { "type": "boolean" },
		"update_disabled": { "type": "boolean" },
		"create_disabled": { "type": "boolean" },
		"delete_disabled": { "type": "boolean" }
	},
	"required": ["type"]
}
`)

type LoginIDKeyConfig struct {
	Key                       string               `json:"key,omitempty"`
	Type                      model.LoginIDKeyType `json:"type,omitempty"`
	MaxLength                 *int                 `json:"max_length,omitempty"`
	Deprecated_ModifyDisabled *bool                `json:"modify_disabled,omitempty"`
	UpdateDisabled            *bool                `json:"update_disabled,omitempty"`
	CreateDisabled            *bool                `json:"create_disabled,omitempty"`
	DeleteDisabled            *bool                `json:"delete_disabled,omitempty"`
}

func (c *LoginIDKeyConfig) SetDefaults() {
	if c.MaxLength == nil {
		switch c.Type {
		case model.LoginIDKeyTypeUsername:
			// Facebook is 50.
			// GitHub is 39.
			// Instagram is 30.
			// Telegram is 32.
			// Seems average is around about ~40 characters.
			c.MaxLength = newInt(40)

		case model.LoginIDKeyTypePhone:
			c.MaxLength = newInt(40)

		default:
			// Maximum length of email address:
			// https://tools.ietf.org/html/rfc3696#section-3
			c.MaxLength = newInt(320)
		}
	}
	if c.Key == "" {
		c.Key = string(c.Type)
	}
	if c.Deprecated_ModifyDisabled == nil {
		c.Deprecated_ModifyDisabled = newBool(false)
	}
	if c.UpdateDisabled == nil {
		c.UpdateDisabled = c.Deprecated_ModifyDisabled
	}
	if c.CreateDisabled == nil {
		c.CreateDisabled = c.Deprecated_ModifyDisabled
	}
	if c.DeleteDisabled == nil {
		c.DeleteDisabled = c.Deprecated_ModifyDisabled
	}
}

var _ = Schema.Add("LoginIDKeyType", `
{
	"type": "string",
	"enum": ["email", "phone", "username"]
}
`)

var _ = Schema.Add("OAuthSSOConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"providers": { "type": "array", "items": { "type": "object" } }
	}
}
`)

type OAuthSSOConfig struct {
	Providers []oauthrelyingparty.ProviderConfig `json:"providers,omitempty"`
}

func (c *OAuthSSOConfig) GetProviderConfig(alias string) (oauthrelyingparty.ProviderConfig, bool) {
	for _, conf := range c.Providers {
		if conf.Alias() == alias {
			cc := conf
			return cc, true
		}
	}
	return nil, false
}

var _ = Schema.Add("PromotionConflictBehavior", `
{
	"type": "string",
	"enum": ["error", "login"]
}
`)

type PromotionConflictBehavior string

const (
	PromotionConflictBehaviorError PromotionConflictBehavior = "error"
	PromotionConflictBehaviorLogin PromotionConflictBehavior = "login"
)

var _ = Schema.Add("IdentityConflictConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"promotion": { "$ref": "#/$defs/PromotionConflictBehavior" }
	}
}
`)

type IdentityConflictConfig struct {
	Promotion PromotionConflictBehavior `json:"promotion,omitempty"`
}

func (c *IdentityConflictConfig) SetDefaults() {
	if c.Promotion == "" {
		c.Promotion = PromotionConflictBehaviorError
	}
}

var _ = Schema.Add("BiometricConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"list_enabled": { "type": "boolean" }
	}
}
`)

type BiometricConfig struct {
	ListEnabled *bool `json:"list_enabled,omitempty"`
}

func (c *BiometricConfig) SetDefaults() {
	if c.ListEnabled == nil {
		c.ListEnabled = newBool(false)
	}
}
