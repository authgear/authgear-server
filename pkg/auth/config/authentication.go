package config

import "github.com/authgear/authgear-server/pkg/core/authn"

var _ = Schema.Add("AuthenticationConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"identities": {
			"type": "array",
			"items": { "$ref": "#/$defs/IdentityType" },
			"uniqueItems": true
		},
		"primary_authenticators": {
			"type": "array",
			"items": { "$ref": "#/$defs/AuthenticatorType" },
			"uniqueItems": true
		},
		"secondary_authenticators": {
			"type": "array",
			"items": { "$ref": "#/$defs/AuthenticatorType" },
			"uniqueItems": true
		},
		"secondary_authentication_mode": { "$ref": "#/$defs/SecondaryAuthenticationMode" }
	}
}
`)

var _ = Schema.Add("IdentityType", `
{
	"type": "string",
	"enum": ["login_id", "oauth", "anonymous"]
}
`)

var _ = Schema.Add("AuthenticatorType", `
{
	"type": "string",
	"enum": ["password", "totp", "oob_otp", "bearer_token"]
}
`)

type AuthenticationConfig struct {
	Identities                  []authn.IdentityType        `json:"identities,omitempty"`
	PrimaryAuthenticators       []authn.AuthenticatorType   `json:"primary_authenticators,omitempty"`
	SecondaryAuthenticators     []authn.AuthenticatorType   `json:"secondary_authenticators,omitempty"`
	SecondaryAuthenticationMode SecondaryAuthenticationMode `json:"secondary_authentication_mode,omitempty"`
}

func (c *AuthenticationConfig) SetDefaults() {
	if len(c.Identities) == 0 {
		c.Identities = []authn.IdentityType{
			authn.IdentityTypeOAuth,
			authn.IdentityTypeLoginID,
		}
	}
	if len(c.PrimaryAuthenticators) == 0 {
		c.PrimaryAuthenticators = []authn.AuthenticatorType{
			authn.AuthenticatorTypePassword,
		}
	}
	if c.SecondaryAuthenticators == nil {
		c.SecondaryAuthenticators = []authn.AuthenticatorType{
			authn.AuthenticatorTypeTOTP,
			authn.AuthenticatorTypeOOB,
			authn.AuthenticatorTypeBearerToken,
		}
	}
	if c.SecondaryAuthenticationMode == "" {
		c.SecondaryAuthenticationMode = SecondaryAuthenticationModeIfExists
	}
}

var _ = Schema.Add("SecondaryAuthenticationMode", `
{
	"type": "string",
	"enum": ["if_requested", "if_exists", "required"]
}
`)

type SecondaryAuthenticationMode string

const (
	SecondaryAuthenticationModeIfRequested SecondaryAuthenticationMode = "if_requested"
	SecondaryAuthenticationModeIfExists    SecondaryAuthenticationMode = "if_exists"
	SecondaryAuthenticationModeRequired    SecondaryAuthenticationMode = "required"
)
