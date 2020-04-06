package config

import "github.com/skygeario/skygear-server/pkg/core/auth/metadata"

//go:generate msgp -tests=false

type IdentityConfiguration struct {
	LoginID *LoginIDConfiguration `json:"login_id,omitempty" yaml:"login_id" msg:"login_id" default_zero_value:"true"`
	OAuth   *OAuthConfiguration   `json:"oauth,omitempty" yaml:"oauth" msg:"oauth" default_zero_value:"true"`
}

type LoginIDConfiguration struct {
	Types *LoginIDTypesConfiguration `json:"types,omitempty" yaml:"types" msg:"types" default_zero_value:"true"`
	Keys  []LoginIDKeyConfiguration  `json:"keys,omitempty" yaml:"keys" msg:"keys"`
}

func (c *LoginIDConfiguration) GetKey(key string) (*LoginIDKeyConfiguration, bool) {
	for _, config := range c.Keys {
		if config.Key == key {
			return &config, true
		}
	}

	return nil, false
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

func (t LoginIDKeyType) IsValid() bool {
	_, validKey := t.MetadataKey()
	return t == LoginIDKeyTypeRaw || validKey
}

type LoginIDTypesConfiguration struct {
	Email    *LoginIDTypeEmailConfiguration    `json:"email,omitempty" yaml:"email" msg:"email" default_zero_value:"true"`
	Username *LoginIDTypeUsernameConfiguration `json:"username,omitempty" yaml:"username" msg:"username" default_zero_value:"true"`
}

type LoginIDTypeEmailConfiguration struct {
	CaseSensitive *bool `json:"case_sensitive" yaml:"case_sensitive" msg:"case_sensitive"`
	BlockPlusSign *bool `json:"block_plus_sign" yaml:"block_plus_sign" msg:"block_plus_sign"`
	IgnoreDotSign *bool `json:"ignore_dot_sign" yaml:"ignore_dot_sign" msg:"ignore_dot_sign"`
}

type LoginIDTypeUsernameConfiguration struct {
	BlockReservedUsernames *bool    `json:"block_reserved_usernames" yaml:"block_reserved_usernames" msg:"block_reserved_usernames"`
	ExcludedKeywords       []string `json:"excluded_keywords,omitempty" yaml:"excluded_keywords" msg:"excluded_keywords"`
	ASCIIOnly              *bool    `json:"ascii_only" yaml:"ascii_only" msg:"ascii_only"`
	CaseSensitive          *bool    `json:"case_sensitive" yaml:"case_sensitive" msg:"case_sensitive"`
}

type LoginIDKeyConfiguration struct {
	Key     string         `json:"key" yaml:"key" msg:"key"`
	Type    LoginIDKeyType `json:"type,omitempty" yaml:"type" msg:"type"`
	Maximum *int           `json:"maximum,omitempty" yaml:"maximum" msg:"maximum"`
}

type OAuthConfiguration struct {
	StateJWTSecret                 string                       `json:"state_jwt_secret,omitempty" yaml:"state_jwt_secret" msg:"state_jwt_secret"`
	ExternalAccessTokenFlowEnabled bool                         `json:"external_access_token_flow_enabled,omitempty" yaml:"external_access_token_flow_enabled" msg:"external_access_token_flow_enabled"`
	Providers                      []OAuthProviderConfiguration `json:"providers,omitempty" yaml:"providers" msg:"providers"`
}

type OAuthProviderType string

const (
	OAuthProviderTypeGoogle    OAuthProviderType = "google"
	OAuthProviderTypeFacebook  OAuthProviderType = "facebook"
	OAuthProviderTypeInstagram OAuthProviderType = "instagram"
	OAuthProviderTypeLinkedIn  OAuthProviderType = "linkedin"
	OAuthProviderTypeAzureADv2 OAuthProviderType = "azureadv2"
	OAuthProviderTypeApple     OAuthProviderType = "apple"
)

type OAuthProviderConfiguration struct {
	ID           string            `json:"id,omitempty" yaml:"id" msg:"id"`
	Type         OAuthProviderType `json:"type,omitempty" yaml:"type" msg:"type"`
	ClientID     string            `json:"client_id,omitempty" yaml:"client_id" msg:"client_id"`
	ClientSecret string            `json:"client_secret,omitempty" yaml:"client_secret" msg:"client_secret"`
	Scope        string            `json:"scope,omitempty" yaml:"scope" msg:"scope"`
	// Tenant is specific to azureadv2
	Tenant string `json:"tenant,omitempty" yaml:"tenant" msg:"tenant"`
	// KeyID and TeamID are specific to apple
	KeyID  string `json:"key_id,omitempty" yaml:"key_id" msg:"key_id"`
	TeamID string `json:"team_id,omitempty" yaml:"team_id" msg:"team_id"`
}
