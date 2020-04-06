package config

//go:generate msgp -tests=false

type IdentityConfiguration struct {
	LoginID *LoginIDConfiguration `json:"login_id,omitempty" yaml:"login_id" msg:"login_id" default_zero_value:"true"`
	OAuth   *OAuthConfiguration   `json:"oauth,omitempty" yaml:"oauth" msg:"oauth" default_zero_value:"true"`
}

type LoginIDConfiguration struct {
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
