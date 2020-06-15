package config

import "github.com/skygeario/skygear-server/pkg/core/authn"

type AuthenticationConfig struct {
	Identities                  []authn.IdentityType        `json:"identities,omitempty"`
	PrimaryAuthenticators       []authn.AuthenticatorType   `json:"primary_authenticators,omitempty"`
	SecondaryAuthenticators     []authn.AuthenticatorType   `json:"secondary_authenticators,omitempty"`
	SecondaryAuthenticationMode SecondaryAuthenticationMode `json:"secondary_authentication_mode,omitempty"`
}

type SecondaryAuthenticationMode string

const (
	SecondaryAuthenticationModeIfRequested SecondaryAuthenticationMode = "if_requested"
	SecondaryAuthenticationModeIfExists    SecondaryAuthenticationMode = "if_exists"
	SecondaryAuthenticationModeRequired    SecondaryAuthenticationMode = "required"
)
