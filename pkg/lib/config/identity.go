package config

import (
	"fmt"
	"strconv"
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
			{Type: LoginIDKeyTypeEmail},
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
		"modify_disabled": { "type": "boolean" }
	},
	"required": ["type"]
}
`)

type LoginIDKeyConfig struct {
	Key            string         `json:"key,omitempty"`
	Type           LoginIDKeyType `json:"type,omitempty"`
	MaxLength      *int           `json:"max_length,omitempty"`
	ModifyDisabled *bool          `json:"modify_disabled,omitempty"`
}

func (c *LoginIDKeyConfig) SetDefaults() {
	if c.MaxLength == nil {
		switch c.Type {
		case LoginIDKeyTypeUsername:
			// Facebook is 50.
			// GitHub is 39.
			// Instagram is 30.
			// Telegram is 32.
			// Seems average is around about ~40 characters.
			c.MaxLength = newInt(40)

		case LoginIDKeyTypePhone:
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
	if c.ModifyDisabled == nil {
		c.ModifyDisabled = newBool(false)
	}
}

var _ = Schema.Add("LoginIDKeyType", `
{
	"type": "string",
	"enum": ["email", "phone", "username"]
}
`)

type LoginIDKeyType string

const (
	LoginIDKeyTypeEmail    LoginIDKeyType = "email"
	LoginIDKeyTypePhone    LoginIDKeyType = "phone"
	LoginIDKeyTypeUsername LoginIDKeyType = "username"
)

var _ = Schema.Add("OAuthSSOConfig", `
{
	"type": "object",
	"additionalProperties": false,
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
		"adfs",
		"apple",
		"wechat"
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
	case OAuthSSOProviderTypeADFS:
		// The supported scopes are observed from a AD FS server.
		return "openid profile email"
	case OAuthSSOProviderTypeApple:
		return "email"
	case OAuthSSOProviderTypeWechat:
		// https://developers.weixin.qq.com/doc/offiaccount/OA_Web_Apps/Wechat_webpage_authorization.html
		return "snsapi_userinfo"
	}

	panic(fmt.Sprintf("oauth: unknown provider type %s", string(t)))
}

const (
	OAuthSSOProviderTypeGoogle    OAuthSSOProviderType = "google"
	OAuthSSOProviderTypeFacebook  OAuthSSOProviderType = "facebook"
	OAuthSSOProviderTypeLinkedIn  OAuthSSOProviderType = "linkedin"
	OAuthSSOProviderTypeAzureADv2 OAuthSSOProviderType = "azureadv2"
	OAuthSSOProviderTypeADFS      OAuthSSOProviderType = "adfs"
	OAuthSSOProviderTypeApple     OAuthSSOProviderType = "apple"
	OAuthSSOProviderTypeWechat    OAuthSSOProviderType = "wechat"
)

var _ = Schema.Add("OAuthSSOWeChatAppType", `
{
	"type": "string",
	"enum": [
		"mobile",
		"web"
	]
}
`)

type OAuthSSOWeChatAppType string

const (
	OAuthSSOWeChatAppTypeWeb    OAuthSSOWeChatAppType = "web"
	OAuthSSOWeChatAppTypeMobile OAuthSSOWeChatAppType = "mobile"
)

var _ = Schema.Add("OAuthSSOProviderConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"alias": { "type": "string" },
		"type": { "$ref": "#/$defs/OAuthSSOProviderType" },
		"modify_disabled": { "type": "boolean" },
		"client_id": { "type": "string" },
		"claims": { "$ref": "#/$defs/VerificationOAuthClaimsConfig" },
		"tenant": { "type": "string" },
		"key_id": { "type": "string" },
		"team_id": { "type": "string" },
		"app_type": { "$ref": "#/$defs/OAuthSSOWeChatAppType" },
		"account_id": { "type": "string", "format": "wechat_account_id"},
		"is_sandbox_account": { "type": "boolean" },
		"wechat_redirect_uris": { "type": "array", "items": { "type": "string", "format": "uri" } },
		"discovery_document_endpoint": { "type": "string", "format": "uri" }
	},
	"required": ["alias", "type", "client_id"],
	"allOf": [
		{
			"if": { "properties": { "type": { "const": "apple" } } },
			"then": {
				"required": ["key_id", "team_id"]
			}
		},
		{
			"if": { "properties": { "type": { "const": "azureadv2" } } },
			"then": {
				"required": ["tenant"]
			}
		},
		{
			"if": { "properties": { "type": { "const": "wechat" } } },
			"then": {
				"required": ["app_type", "account_id"]
			}
		},
		{
			"if": { "properties": { "type": { "const": "adfs" } } },
			"then": {
				"required": ["discovery_document_endpoint"]
			}
		}
	]
}
`)

type OAuthSSOProviderConfig struct {
	Alias          string                         `json:"alias,omitempty"`
	Type           OAuthSSOProviderType           `json:"type,omitempty"`
	ModifyDisabled *bool                          `json:"modify_disabled,omitempty"`
	ClientID       string                         `json:"client_id,omitempty"`
	Claims         *VerificationOAuthClaimsConfig `json:"claims,omitempty"`

	// Tenant is specific to `azureadv2`
	Tenant string `json:"tenant,omitempty"`

	// KeyID and TeamID are specific to `apple`
	KeyID  string `json:"key_id,omitempty"`
	TeamID string `json:"team_id,omitempty"`

	// AppType is specific to `wechat`, support web or mobile
	AppType            OAuthSSOWeChatAppType `json:"app_type,omitempty"`
	AccountID          string                `json:"account_id,omitempty"`
	IsSandboxAccount   bool                  `json:"is_sandbox_account,omitempty"`
	WeChatRedirectURIs []string              `json:"wechat_redirect_uris,omitempty"`

	// DiscoveryDocumentEndpoint is specific to `adfs`.
	DiscoveryDocumentEndpoint string `json:"discovery_document_endpoint,omitempty"`
}

func (c *OAuthSSOProviderConfig) SetDefaults() {
	if c.ModifyDisabled == nil {
		c.ModifyDisabled = newBool(false)
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
	case OAuthSSOProviderTypeWechat:
		// WeChat does NOT support OIDC.
		// In the same Weixin Open Platform account, the user UnionID is unique.
		// The id is scoped to Open Platform account.
		// https://developers.weixin.qq.com/miniprogram/en/dev/framework/open-ability/union-id.html
		keys["account_id"] = c.AccountID
		keys["is_sandbox_account"] = strconv.FormatBool(c.IsSandboxAccount)
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

func NewProviderID(claims map[string]interface{}) ProviderID {
	id := ProviderID{Keys: map[string]interface{}{}}
	for k, v := range claims {
		if k == "type" {
			id.Type = v.(string)
		} else {
			id.Keys[k] = v.(string)
		}
	}
	return id
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
