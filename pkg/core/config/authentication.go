package config

//go:generate msgp -tests=false
type AuthenticationConfiguration struct {
	Identities                  []string                    `json:"identities,omitempty" yaml:"identities" msg:"identities"`
	PrimaryAuthenticators       []string                    `json:"primary_authenticators,omitempty" yaml:"primary_authenticators" msg:"primary_authenticators"`
	SecondaryAuthenticators     []string                    `json:"secondary_authenticators" yaml:"secondary_authenticators" msg:"secondary_authenticators"`
	SecondaryAuthenticationMode SecondaryAuthenticationMode `json:"secondary_authentication_mode,omitempty" yaml:"secondary_authentication_mode" msg:"secondary_authentication_mode"`
	Secret                      string                      `json:"secret,omitempty" yaml:"secret" msg:"secret"`
}

type SecondaryAuthenticationMode string

const (
	SecondaryAuthenticationModeIfRequested SecondaryAuthenticationMode = "if_requested"
	SecondaryAuthenticationModeIfExists    SecondaryAuthenticationMode = "if_exists"
	SecondaryAuthenticationModeRequired    SecondaryAuthenticationMode = "required"
)
