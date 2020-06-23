package config

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
)

var _ = Schema.Add("IdentityConfig", `
{
	"type": "object",
	"properties": {
		"login_id": { "$ref": "#/$defs/LoginIDConfig" },
		"oauth": { "$ref": "#/$defs/OAuthSSOConfig" },
		"on_conflict": { "$ref": "#/$defs/IdentityConflictConfig" }
	}
}
`)

type IdentityConfig struct {
	LoginID    *LoginIDConfig          `json:"login_id,omitempty"`
	OAuth      *OAuthSSOConfig         `json:"oauth,omitempty"`
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

func (c *LoginIDConfig) SetDefaults() {
	if len(c.Keys) == 0 {
		c.Keys = []LoginIDKeyConfig{
			{Key: "email", Type: LoginIDKeyType(metadata.Email), Maximum: newInt(1)},
			{Key: "phone", Type: LoginIDKeyType(metadata.Phone), Maximum: newInt(1)},
			{Key: "username", Type: LoginIDKeyType(metadata.Username), Maximum: newInt(1)},
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

func (c *LoginIDUsernameConfig) SetDefaults() {
	if c.BlockReservedUsernames == nil {
		c.BlockReservedUsernames = newBool(true)
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

func (c *LoginIDKeyConfig) SetDefaults() {
	if c.Maximum == nil {
		c.Maximum = newInt(1)
	}
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

var _ = Schema.Add("OAuthSSOConfig", `
{
	"type": "object",
	"properties": {
		"providers": { "type": "array", "items": { "$ref": "#/$defs/OAuthSSOProviderConfig" } }
	}
}
`)

type OAuthSSOConfig struct {
	Providers []OAuthSSOProviderConfig `json:"providers,omitempty"`
}

func (c *OAuthSSOConfig) GetProviderConfig(alias string) (*OAuthSSOProviderConfig, bool) {
	for _, conf := range c.Providers {
		if conf.Alias == alias {
			cc := conf
			return &cc, true
		}
	}
	return nil, false
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

func (t OAuthSSOProviderType) Scope() string {
	switch t {
	case OAuthSSOProviderTypeGoogle:
		// https://developers.google.com/identity/protocols/googlescopes#google_sign-in
		return "openid profile email"
	case OAuthSSOProviderTypeFacebook:
		// https://developers.facebook.com/docs/facebook-login/permissions/#reference-default
		// https://developers.facebook.com/docs/facebook-login/permissions/#reference-email
		return "email"
	case OAuthSSOProviderTypeLinkedIn:
		// https://docs.microsoft.com/en-us/linkedin/shared/integrations/people/profile-api?context=linkedin/compliance/context
		// https://docs.microsoft.com/en-us/linkedin/shared/integrations/people/primary-contact-api?context=linkedin/compliance/context
		return "r_liteprofile r_emailaddress"
	case OAuthSSOProviderTypeAzureADv2:
		// https://docs.microsoft.com/en-us/azure/active-directory/develop/v2-permissions-and-consent#openid-connect-scopes
		return "openid profile email"
	case OAuthSSOProviderTypeApple:
		return "email"
	}

	panic(fmt.Sprintf("oauth: unknown provider type %s", string(t)))
}

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

func (c *OAuthSSOProviderConfig) SetDefaults() {
	if c.Alias == "" {
		c.Alias = string(c.Type)
	}
}

func (c *OAuthSSOProviderConfig) ProviderID() ProviderID {
	keys := map[string]interface{}{}
	switch c.Type {
	case OAuthSSOProviderTypeGoogle:
		// Google supports OIDC.
		// sub is public, not scoped to anything so changing client_id does not affect sub.
		// Therefore, ProviderID is simply Type.
		//
		// Rotating the OAuth application is OK.
		break
	case OAuthSSOProviderTypeFacebook:
		// Facebook does NOT support OIDC.
		// Facebook user ID is scoped to client_id.
		// Therefore, ProviderID is Type + client_id.
		//
		// Rotating the OAuth application is problematic.
		// But if email remains unchanged, the user can associate their account.
		keys["client_id"] = c.ClientID
	case OAuthSSOProviderTypeLinkedIn:
		// LinkedIn is the same as Facebook.
		keys["client_id"] = c.ClientID
	case OAuthSSOProviderTypeAzureADv2:
		// Azure AD v2 supports OIDC.
		// sub is pairwise and is scoped to client_id.
		// However, oid is powerful alternative to sub.
		// oid is also pairwise and is scoped to tenant.
		// We use oid as ProviderSubjectID so ProviderID is Type + tenant.
		//
		// Rotating the OAuth application is OK.
		// But rotating the tenant is problematic.
		// But if email remains unchanged, the user can associate their account.
		keys["tenant"] = c.Tenant
	case OAuthSSOProviderTypeApple:
		// Apple supports OIDC.
		// sub is pairwise and is scoped to team_id.
		// Therefore, ProviderID is Type + team_id.
		//
		// Rotating the OAuth application is OK.
		// But rotating the Apple Developer account is problematic.
		// Since Apple has private relay to hide the real email,
		// the user may not be associate their account.
		keys["team_id"] = c.TeamID
	}

	return ProviderID{
		Type: string(c.Type),
		Keys: keys,
	}
}

// ProviderID combining with a subject ID identifies an user from an external system.
type ProviderID struct {
	Type string
	Keys map[string]interface{}
}

func (p ProviderID) Claims() map[string]interface{} {
	claim := map[string]interface{}{}
	claim["type"] = p.Type
	for k, v := range p.Keys {
		claim[k] = v
	}
	return claim
}

func (p ProviderID) Equal(that *ProviderID) bool {
	if p.Type != that.Type || len(p.Keys) != len(that.Keys) {
		return false
	}
	for k, v := range p.Keys {
		if tv, ok := that.Keys[k]; !ok || tv != v {
			return false
		}
	}
	return true
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

func (c *IdentityConflictConfig) SetDefaults() {
	if c.Promotion == "" {
		c.Promotion = PromotionConflictBehaviorError
	}
}
