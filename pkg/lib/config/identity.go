package config

import (
	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var _ = Schema.Add("IdentityConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"ldap": { "$ref": "#/$defs/LDAPConfig" },
		"login_id": { "$ref": "#/$defs/LoginIDConfig" },
		"oauth": { "$ref": "#/$defs/OAuthSSOConfig" },
		"biometric": { "$ref": "#/$defs/BiometricConfig" },
		"on_conflict": { "$ref": "#/$defs/IdentityConflictConfig" }
	}
}
`)

type IdentityConfig struct {
	LDAP       *LDAPConfig             `json:"ldap,omitempty"`
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
	// NOTE(tung): omitempty cannot be applied to `keys`, because empty array is a valid config.
	// When the array is empty and omitempty is set, the key is omitted in the output of portal graphql api.
	// As a result, when portal save the config, `keys` will be undefined, and the value in go will be nil.
	// And when the value is nil, default will be applied and added email to keys, which is unexpected.
	// Use `omitzero` instead so empty array will be outputted, while nil will still be omitted.
	Keys []LoginIDKeyConfig `json:"keys,omitzero"`
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
		"block_free_email_provider_domains" : {"type": "boolean"},
		"block_disposable_email_domains": {"type": "boolean"}
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
					"domain_allowlist_enabled": { "enum": [false] }
				}
			}
		},
		{
			"if": {
				"properties": {
					"block_disposable_email_domains": { "enum": [true] }
				},
				"required": ["block_disposable_email_domains"]
			},
			"then": {
				"properties": {
					"domain_allowlist_enabled": { "enum": [false] }
				}
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
	BlockDisposableEmailDomains   *bool `json:"block_disposable_email_domains,omitempty"`
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
	if c.BlockDisposableEmailDomains == nil {
		c.BlockDisposableEmailDomains = newBool(false)
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
		b := *c.Deprecated_ModifyDisabled
		c.UpdateDisabled = &b
	}
	if c.CreateDisabled == nil {
		b := *c.Deprecated_ModifyDisabled
		c.CreateDisabled = &b
	}
	if c.DeleteDisabled == nil {
		b := *c.Deprecated_ModifyDisabled
		c.DeleteDisabled = &b
	}

	c.Deprecated_ModifyDisabled = nil
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

type OAuthSSOProviderCredentialsBehavior string

const (
	OAuthSSOProviderCredentialsBehaviorUseProjectCredentials OAuthSSOProviderCredentialsBehavior = "use_project_credentials"
	//nolint: gosec
	OAuthSSOProviderCredentialsBehaviorUseDemoCredentials OAuthSSOProviderCredentialsBehavior = "use_demo_credentials"
)

func OAuthSSOProviderConfigSchemaBuilder(providerSchemaBuilder validation.SchemaBuilder) validation.SchemaBuilder {
	builder := validation.SchemaBuilder{}
	builder.Properties().
		Property("alias", validation.SchemaBuilder{}.Type(validation.TypeString).MinLength(1)).
		Property("modify_disabled", validation.SchemaBuilder{}.Type(validation.TypeBoolean)).
		Property("create_disabled", validation.SchemaBuilder{}.Type(validation.TypeBoolean)).
		Property("delete_disabled", validation.SchemaBuilder{}.Type(validation.TypeBoolean)).
		Property("do_not_store_identity_attributes", validation.SchemaBuilder{}.Type(validation.TypeBoolean)).
		Property("include_identity_attributes_in_id_token", validation.SchemaBuilder{}.Type(validation.TypeBoolean)).
		Property("credentials_behavior", validation.SchemaBuilder{}.Type(validation.TypeString).Enum("use_project_credentials", "use_demo_credentials"))
	builder.AddRequired("alias")

	_if := validation.SchemaBuilder{}
	_if.Properties().
		Property("credentials_behavior", validation.SchemaBuilder{}.Const("use_demo_credentials"))
	_if.Required("credentials_behavior")

	builder.AllOf(validation.SchemaBuilder{}.If(_if).
		Else(providerSchemaBuilder))

	return builder
}

type OAuthSSOProviderConfig oauthrelyingparty.ProviderConfig

func (c OAuthSSOProviderConfig) SetDefaults() {
	if _, ok := c["modify_disabled"].(bool); !ok {
		c["modify_disabled"] = false
	}

	if _, ok := c["create_disabled"].(bool); !ok {
		c["create_disabled"] = c["modify_disabled"].(bool)
	}

	if _, ok := c["delete_disabled"].(bool); !ok {
		c["delete_disabled"] = c["modify_disabled"].(bool)
	}

	// Intentionally not setting default of do_not_store_identity_attributes.
	// Intentionally not setting default of include_identity_attributes_in_id_token.

	c.AsProviderConfig().SetDefaults()

	// Cleanup deprecated fields
	delete(c, "modify_disabled")
}

func (c OAuthSSOProviderConfig) AsProviderConfig() oauthrelyingparty.ProviderConfig {
	return oauthrelyingparty.ProviderConfig(c)
}

func (c OAuthSSOProviderConfig) Alias() string {
	alias, ok := c["alias"].(string)
	if ok {
		return alias
	}
	// This method is called in validateOAuthProvider which is part of the validation process
	// So it is possible that alias is an invalid value
	return ""
}

func (c OAuthSSOProviderConfig) CreateDisabled() bool {
	return c["create_disabled"].(bool)
}
func (c OAuthSSOProviderConfig) DeleteDisabled() bool {
	return c["delete_disabled"].(bool)
}
func (c OAuthSSOProviderConfig) GetCredentialsBehavior() OAuthSSOProviderCredentialsBehavior {
	v, ok := c["credentials_behavior"].(string)
	if !ok {
		return OAuthSSOProviderCredentialsBehaviorUseProjectCredentials
	}
	return OAuthSSOProviderCredentialsBehavior(v)
}

func (c OAuthSSOProviderConfig) DoNotStoreIdentityAttributes() bool {
	b, ok := c["do_not_store_identity_attributes"].(bool)
	if !ok {
		// If absent, the default is false, which means DO store identity attributes.
		// That is the original behavior.
		return false
	}
	return b
}

func (c OAuthSSOProviderConfig) IncludeIdentityAttributesInIDToken() bool {
	b, ok := c["include_identity_attributes_in_id_token"].(bool)
	if !ok {
		return false
	}
	return b
}

type OAuthProviderStatus string

const (
	OAuthProviderStatusActive               OAuthProviderStatus = "active"
	OAuthProviderStatusMissingCredentials   OAuthProviderStatus = "missing_credentials"
	OAuthProviderStatusUsingDemoCredentials OAuthProviderStatus = "using_demo_credentials" // nolint: gosec
)

func (c OAuthSSOProviderConfig) ComputeProviderStatus(demoCredentials *SSOOAuthDemoCredentials) OAuthProviderStatus {
	if c.GetCredentialsBehavior() == OAuthSSOProviderCredentialsBehaviorUseProjectCredentials {
		return OAuthProviderStatusActive
	}
	if demoCredentials == nil {
		return OAuthProviderStatusMissingCredentials
	}
	typ := c.AsProviderConfig().Type()
	_, ok := demoCredentials.LookupByProviderType(typ)
	if ok {
		return OAuthProviderStatusUsingDemoCredentials
	}
	return OAuthProviderStatusMissingCredentials
}

type OAuthSSOConfig struct {
	Providers []OAuthSSOProviderConfig `json:"providers,omitempty"`
}

func (c *OAuthSSOConfig) GetProviderConfig(alias string) (oauthrelyingparty.ProviderConfig, bool) {
	for _, conf := range c.Providers {
		if conf.Alias() == alias {
			cc := conf
			return cc.AsProviderConfig(), true
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
