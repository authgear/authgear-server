package config

import "github.com/skygeario/skygear-server/pkg/core/authn"

var _ = Schema.Add("AuthenticationConfig", `
{
	"type": "object",
	"properties": {
		"identities": { "type": "array", "items": { "$ref": "#/$defs/IdentityType" } },
		"primary_authenticators": { "type": "array", "items": { "$ref": "#/$defs/AuthenticatorType" } },
		"secondary_authenticators": { "type": "array", "items": { "$ref": "#/$defs/AuthenticatorType" } },
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
	"enum": ["password", "totp", "oob_otp", "recovery_code", "bearer_token"]
}
`)

type AuthenticationConfig struct {
	Identities                  []authn.IdentityType        `json:"identities,omitempty"`
	PrimaryAuthenticators       []authn.AuthenticatorType   `json:"primary_authenticators,omitempty"`
	SecondaryAuthenticators     []authn.AuthenticatorType   `json:"secondary_authenticators,omitempty"`
	SecondaryAuthenticationMode SecondaryAuthenticationMode `json:"secondary_authentication_mode,omitempty"`
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
