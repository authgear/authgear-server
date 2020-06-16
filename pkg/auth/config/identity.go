package config

import "github.com/skygeario/skygear-server/pkg/core/auth/metadata"

var _ = Schema.Add("IdentityConfig", `
{
	"type": "object",
	"properties": {
		"login_id": { "$ref": "#/$defs/LoginIDConfig" },
		"sso": { "$ref": "#/$defs/SSOConfig" },
		"sso": { "$ref": "#/$defs/IdentityConflictConfig" }
	}
}
`)

type IdentityConfig struct {
	LoginID    *LoginIDConfig          `json:"login_id,omitempty"`
	SSO        *SSOConfig              `json:"sso,omitempty"`
	OnConflict *IdentityConflictConfig `json:"on_conflict,omitempty"`
}

var _ = Schema.Add("LoginIDConfig", `
{
	"type": "object",
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
	"properties": {
		"case_sensitive": { "type": "boolean" },
		"block_plus_sign": { "type": "boolean" },
		"ignore_dot_sign": { "type": "boolean" }
	}
}
`)

type LoginIDEmailConfig struct {
	CaseSensitive *bool `json:"case_sensitive,omitempty"`
	BlockPlusSign *bool `json:"block_plus_sign,omitempty"`
	IgnoreDotSign *bool `json:"ignore_dot_sign,omitempty"`
}

var _ = Schema.Add("LoginIDUsernameConfig", `
{
	"type": "object",
	"properties": {
		"block_reserved_usernames": { "type": "boolean" },
		"excluded_keywords": { "type": "array", "items": { "type": "string" } },
		"ascii_only": { "type": "boolean" },
		"case_sensitive": { "type": "boolean" }
	}
}
`)

type LoginIDUsernameConfig struct {
	BlockReservedUsernames *bool    `json:"block_reserved_usernames,omitempty"`
	ExcludedKeywords       []string `json:"excluded_keywords,omitempty"`
	ASCIIOnly              *bool    `json:"ascii_only,omitempty"`
	CaseSensitive          *bool    `json:"case_sensitive,omitempty"`
}

var _ = Schema.Add("LoginIDKeyConfig", `
{
	"type": "object",
	"properties": {
		"key": { "type": "string" },
		"type": { "$ref": "#/$defs/LoginIDKeyType" },
		"maximum": { "type": "integer" }
	},
	"required": ["key", "type"]
}
`)

type LoginIDKeyConfig struct {
	Key     string         `json:"key,omitempty"`
	Type    LoginIDKeyType `json:"type,omitempty"`
	Maximum *int           `json:"maximum,omitempty"`
}

var _ = Schema.Add("LoginIDKeyType", `
{
	"type": "string",
	"enum": ["raw", "email", "phone", "username"]
}
`)

type LoginIDKeyType string

const LoginIDKeyTypeRaw LoginIDKeyType = "raw"

func (t LoginIDKeyType) MetadataKey() (metadata.StandardKey, bool) {
	for _, key := range metadata.AllKeys() {
		if string(t) == string(key) {
			return key, true
		}
	}
	return "", false
}

var _ = Schema.Add("SSOConfig", `
{
	"type": "object",
	"properties": {
		"oauth_providers": { "type": "array", "items": { "$ref": "#/$defs/OAuthSSOProviderConfig" } }
	}
}
`)

type SSOConfig struct {
	OAuthProviders []OAuthSSOProviderConfig `json:"oauth_providers,omitempty"`
}

var _ = Schema.Add("OAuthSSOProviderType", `
{
	"type": "string",
	"enum": [
		"google",
		"facebook",
		"linkedin",
		"azureadv2",
		"apple"
	]
}
`)

type OAuthSSOProviderType string

const (
	OAuthSSOProviderTypeGoogle    OAuthSSOProviderType = "google"
	OAuthSSOProviderTypeFacebook  OAuthSSOProviderType = "facebook"
	OAuthSSOProviderTypeLinkedIn  OAuthSSOProviderType = "linkedin"
	OAuthSSOProviderTypeAzureADv2 OAuthSSOProviderType = "azureadv2"
	OAuthSSOProviderTypeApple     OAuthSSOProviderType = "apple"
)

var _ = Schema.Add("OAuthSSOProviderConfig", `
{
	"type": "object",
	"properties": {
		"alias": { "type": "string" },
		"type": { "$ref": "#/$defs/OAuthSSOProviderType" },
		"client_id": { "type": "string" },
		"tenant": { "type": "string" },
		"key_id": { "type": "string" },
		"team_id": { "type": "string" }
	},
	"required": ["type", "client_id"]
}
`)

type OAuthSSOProviderConfig struct {
	Alias    string               `json:"alias,omitempty"`
	Type     OAuthSSOProviderType `json:"type,omitempty"`
	ClientID string               `json:"client_id,omitempty"`

	// Tenant is specific to azureadv2
	Tenant string `json:"tenant,omitempty"`

	// KeyID and TeamID are specific to apple
	KeyID  string `json:"key_id,omitempty"`
	TeamID string `json:"team_id,omitempty"`
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
	"properties": {
		"promotion": { "$ref": "#/$defs/PromotionConflictBehavior" }
	}
}
`)

type IdentityConflictConfig struct {
	Promotion PromotionConflictBehavior `json:"promotion,omitempty"`
}
