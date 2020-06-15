package config

import "github.com/skygeario/skygear-server/pkg/core/auth/metadata"

type IdentityConfig struct {
	LoginID    *LoginIDConfig          `json:"login_id,omitempty"`
	SSO        *SSOConfig              `json:"sso,omitempty"`
	OnConflict *IdentityConflictConfig `json:"on_conflict,omitempty"`
}

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

type LoginIDTypesConfig struct {
	Email    *LoginIDEmailConfig    `json:"email,omitempty"`
	Username *LoginIDUsernameConfig `json:"username,omitempty"`
}

type LoginIDEmailConfig struct {
	CaseSensitive *bool `json:"case_sensitive,omitempty"`
	BlockPlusSign *bool `json:"block_plus_sign,omitempty"`
	IgnoreDotSign *bool `json:"ignore_dot_sign,omitempty"`
}

type LoginIDUsernameConfig struct {
	BlockReservedUsernames *bool    `json:"block_reserved_usernames,omitempty"`
	ExcludedKeywords       []string `json:"excluded_keywords,omitempty"`
	ASCIIOnly              *bool    `json:"ascii_only,omitempty"`
	CaseSensitive          *bool    `json:"case_sensitive,omitempty"`
}

type LoginIDKeyConfig struct {
	Key     string         `json:"key,omitempty"`
	Type    LoginIDKeyType `json:"type,omitempty"`
	Maximum *int           `json:"maximum,omitempty"`
}

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

type SSOConfig struct {
	OAuthProviders []OAuthSSOProviderConfig `json:"oauth_providers,omitempty"`
}

type OAuthSSOProviderType string

const (
	OAuthSSOProviderTypeGoogle    OAuthSSOProviderType = "google"
	OAuthSSOProviderTypeFacebook  OAuthSSOProviderType = "facebook"
	OAuthSSOProviderTypeLinkedIn  OAuthSSOProviderType = "linkedin"
	OAuthSSOProviderTypeAzureADv2 OAuthSSOProviderType = "azureadv2"
	OAuthSSOProviderTypeApple     OAuthSSOProviderType = "apple"
)

type OAuthSSOProviderConfig struct {
	ID       string               `json:"id,omitempty"`
	Type     OAuthSSOProviderType `json:"type,omitempty"`
	ClientID string               `json:"client_id,omitempty"`

	// Tenant is specific to azureadv2
	Tenant string `json:"tenant,omitempty"`

	// KeyID and TeamID are specific to apple
	KeyID  string `json:"key_id,omitempty"`
	TeamID string `json:"team_id,omitempty"`
}

type IdentityConflictConfig struct {
	Promotion PromotionConflictBehavior `json:"promotion,omitempty"`
}

type PromotionConflictBehavior string

const (
	PromotionConflictBehaviorError PromotionConflictBehavior = "error"
	PromotionConflictBehaviorLogin PromotionConflictBehavior = "login"
)
