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
			"items": { "$ref": "#/$defs/PrimaryAuthenticatorType" },
			"uniqueItems": true
		},
		"secondary_authenticators": {
			"type": "array",
			"items": { "$ref": "#/$defs/SecondaryAuthenticatorType" },
			"uniqueItems": true
		},
		"secondary_authentication_mode": { "$ref": "#/$defs/SecondaryAuthenticationMode" },
		"device_token": { "$ref": "#/$defs/DeviceTokenConfig" },
		"recovery_code": { "$ref": "#/$defs/RecoveryCodeConfig" }
	}
}
`)

var _ = Schema.Add("IdentityType", `
{
	"type": "string",
	"enum": ["login_id", "oauth", "anonymous"]
}
`)

var _ = Schema.Add("PrimaryAuthenticatorType", `
{
	"type": "string",
	"enum": ["password", "oob_otp"]
}
`)

var _ = Schema.Add("SecondaryAuthenticatorType", `
{
	"type": "string",
	"enum": ["password", "oob_otp", "totp"]
}
`)

type AuthenticationConfig struct {
	Identities                  []authn.IdentityType        `json:"identities,omitempty"`
	PrimaryAuthenticators       []authn.AuthenticatorType   `json:"primary_authenticators,omitempty"`
	SecondaryAuthenticators     []authn.AuthenticatorType   `json:"secondary_authenticators,omitempty"`
	SecondaryAuthenticationMode SecondaryAuthenticationMode `json:"secondary_authentication_mode,omitempty"`
	DeviceToken                 *DeviceTokenConfig          `json:"device_token,omitempty"`
	RecoveryCode                *RecoveryCodeConfig         `json:"recovery_code,omitempty"`
}

func (c *AuthenticationConfig) SetDefaults() {
	if c.Identities == nil {
		c.Identities = []authn.IdentityType{
			authn.IdentityTypeOAuth,
			authn.IdentityTypeLoginID,
		}
	}
	if c.PrimaryAuthenticators == nil {
		c.PrimaryAuthenticators = []authn.AuthenticatorType{
			authn.AuthenticatorTypePassword,
		}
	}
	if c.SecondaryAuthenticators == nil {
		c.SecondaryAuthenticators = []authn.AuthenticatorType{
			authn.AuthenticatorTypeTOTP,
			authn.AuthenticatorTypeOOB,
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

var _ = Schema.Add("DeviceTokenConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"disabled": { "type": "boolean" },
		"expire_in_days": { "$ref": "#/$defs/DurationDays" }
	}
}
`)

type DeviceTokenConfig struct {
	Disabled bool         `json:"disabled,omitempty"`
	ExpireIn DurationDays `json:"expire_in_days,omitempty"`
}

func (c *DeviceTokenConfig) SetDefaults() {
	if c.ExpireIn == 0 {
		c.ExpireIn = DurationDays(30)
	}
}

var _ = Schema.Add("RecoveryCodeConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"count": { "type": "integer" },
		"list_enabled": { "type": "boolean" }
	}
}
`)

type RecoveryCodeConfig struct {
	Count       int  `json:"count,omitempty"`
	ListEnabled bool `json:"list_enabled,omitempty"`
}

func (c *RecoveryCodeConfig) SetDefaults() {
	if c.Count == 0 {
		c.Count = 16
	}
}
