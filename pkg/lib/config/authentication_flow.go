package config

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
)

var _ = Schema.Add("AuthenticationFlowConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"signup_flows": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/AuthenticationFlowSignupFlow" }
		},
		"login_flows": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/AuthenticationFlowLoginFlow" }
		},
		"signup_login_flows": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/AuthenticationFlowSignupLoginFlow" }
		},
		"reauth_flows": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/AuthenticationFlowReauthFlow" }
		}
	}
}
`)

var _ = Schema.Add("AuthenticationFlowObjectID", `
{
	"type": "string",
	"pattern": "^[a-zA-Z_][a-zA-Z0-9_]*$"
}
`)

var _ = Schema.Add("AuthenticationFlowIdentificationMethod", `
{
	"type": "string",
	"enum": [
		"email",
		"phone",
		"username",
		"oauth",
		"passkey",
		"siwe"
	]
}
`)

var _ = Schema.Add("AuthenticationFlowSignupFlow", `
{
	"type": "object",
	"required": ["id", "steps"],
	"properties": {
		"id": { "$ref": "#/$defs/AuthenticationFlowObjectID" },
		"steps": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/AuthenticationFlowSignupFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("AuthenticationFlowSignupFlowStep", `
{
	"type": "object",
	"required": ["type"],
	"properties": {
		"id": { "$ref": "#/$defs/AuthenticationFlowObjectID" },
		"type": {
			"type": "string",
			"enum": [
				"identify",
				"authenticate",
				"verify",
				"user_profile",
				"recovery_code"
			]
		}
	},
	"allOf": [
		{
			"if": {
				"required": ["type"],
				"properties": {
					"type": { "const": "identify" }
				}
			},
			"then": {
				"required": ["one_of"],
				"properties": {
					"one_of": {
						"type": "array",
						"items": { "$ref": "#/$defs/AuthenticationFlowSignupFlowIdentify" }
					}
				}
			}
		},
		{
			"if": {
				"required": ["type"],
				"properties": {
					"type": { "const": "authenticate" }
				}
			},
			"then": {
				"required": ["one_of"],
				"properties": {
					"one_of": {
						"type": "array",
						"items": { "$ref": "#/$defs/AuthenticationFlowSignupFlowAuthenticate" }
					}
				}
			}
		},
		{
			"if": {
				"required": ["type"],
				"properties": {
					"type": { "const": "verify" }
				}
			},
			"then": {
				"required": ["target_step"],
				"properties": {
					"target_step": { "$ref": "#/$defs/AuthenticationFlowObjectID" }
				}
			}
		},
		{
			"if": {
				"required": ["type"],
				"properties": {
					"type": { "const": "user_profile" }
				}
			},
			"then": {
				"required": ["user_profile"],
				"properties": {
					"user_profile": {
						"type": "array",
						"items": { "$ref": "#/$defs/AuthenticationFlowSignupFlowUserProfile" }
					}
				}
			}
		}
	]
}
`)

var _ = Schema.Add("AuthenticationFlowSignupFlowIdentify", `
{
	"type": "object",
	"required": ["identification"],
	"properties": {
		"identification": { "$ref": "#/$defs/AuthenticationFlowIdentificationMethod" },
		"steps": {
			"type": "array",
			"items": { "$ref": "#/$defs/AuthenticationFlowSignupFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("AuthenticationFlowSignupFlowAuthenticate", `
{
	"type": "object",
	"required": ["authentication"],
	"properties": {
		"authentication": {
			"type": "string",
			"enum": [
				"primary_password",
				"primary_passkey",
				"primary_oob_otp_email",
				"primary_oob_otp_sms",
				"secondary_password",
				"secondary_totp",
				"secondary_oob_otp_email",
				"secondary_oob_otp_sms"
			]
		},
		"target_step": { "$ref": "#/$defs/AuthenticationFlowObjectID" },
		"steps": {
			"type": "array",
			"items": { "$ref": "#/$defs/AuthenticationFlowSignupFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("AuthenticationFlowSignupFlowUserProfile", `
{
	"type": "object",
	"required": ["pointer", "required"],
	"properties": {
		"pointer": {
			"type": "string",
			"format": "json-pointer"
		},
		"required": { "type": "boolean" }
	}
}
`)

var _ = Schema.Add("AuthenticationFlowLoginFlow", `
{
	"type": "object",
	"required": ["id", "steps"],
	"properties": {
		"id": { "$ref": "#/$defs/AuthenticationFlowObjectID" },
		"steps": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/AuthenticationFlowLoginFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("AuthenticationFlowLoginFlowStep", `
{
	"type": "object",
	"required": ["type"],
	"properties": {
		"id": { "$ref": "#/$defs/AuthenticationFlowObjectID" },
		"type": {
			"type": "string",
			"enum": [
				"identify",
				"authenticate",
				"change_password"
			]
		}
	},
	"allOf": [
		{
			"if": {
				"required": ["type"],
				"properties": {
					"type": { "const": "identify" }
				}
			},
			"then": {
				"required": ["one_of"],
				"properties": {
					"one_of": {
						"type": "array",
						"items": { "$ref": "#/$defs/AuthenticationFlowLoginFlowIdentify" }
					}
				}
			}
		},
		{
			"if": {
				"required": ["type"],
				"properties": {
					"type": { "const": "authenticate" }
				}
			},
			"then": {
				"required": ["one_of"],
				"properties": {
					"optional": { "type": "boolean" },
					"one_of": {
						"type": "array",
						"items": { "$ref": "#/$defs/AuthenticationFlowLoginFlowAuthenticate" }
					}
				}
			}
		},
		{
			"if": {
				"required": ["type"],
				"properties": {
					"type": { "const": "change_password" }
				}
			},
			"then": {
				"required": ["target_step"],
				"properties": {
					"target_step": { "$ref": "#/$defs/AuthenticationFlowObjectID" }
				}
			}
		}
	]
}
`)

var _ = Schema.Add("AuthenticationFlowLoginFlowIdentify", `
{
	"type": "object",
	"required": ["identification"],
	"properties": {
		"identification": { "$ref": "#/$defs/AuthenticationFlowIdentificationMethod" },
		"steps": {
			"type": "array",
			"items": { "$ref": "#/$defs/AuthenticationFlowLoginFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("AuthenticationFlowLoginFlowAuthenticate", `
{
	"type": "object",
	"required": ["authentication"],
	"properties": {
		"authentication": {
			"type": "string",
			"enum": [
				"primary_password",
				"primary_passkey",
				"primary_oob_otp_email",
				"primary_oob_otp_sms",
				"secondary_password",
				"secondary_totp",
				"secondary_oob_otp_email",
				"secondary_oob_otp_sms",
				"recovery_code",
				"device_token"
			]
		},
		"target_step": { "$ref": "#/$defs/AuthenticationFlowObjectID" },
		"steps": {
			"type": "array",
			"items": { "$ref": "#/$defs/AuthenticationFlowLoginFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("AuthenticationFlowSignupLoginFlow", `
{
	"type": "object",
	"required": ["id", "steps"],
	"properties": {
		"id": { "$ref": "#/$defs/AuthenticationFlowObjectID" },
		"steps": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/AuthenticationFlowSignupLoginFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("AuthenticationFlowSignupLoginFlowStep", `
{
	"type": "object",
	"required": ["type"],
	"properties": {
		"id": { "$ref": "#/$defs/AuthenticationFlowObjectID" },
		"type": {
			"type": "string",
			"enum": ["identify"]
		}
	},
	"allOf": [
		{
			"if": {
				"required": ["type"],
				"properties": {
					"type": { "const": "identify" }
				}
			},
			"then": {
				"required": ["one_of"],
				"properties": {
					"one_of": {
						"type": "array",
						"items": { "$ref": "#/$defs/AuthenticationFlowSignupLoginFlowIdentify" }
					}
				}
			}
		}
	]
}
`)

var _ = Schema.Add("AuthenticationFlowSignupLoginFlowIdentify", `
{
	"type": "object",
	"required": ["identification", "signup_flow", "login_flow"],
	"properties": {
		"identification": { "$ref": "#/$defs/AuthenticationFlowIdentificationMethod" },
		"signup_flow": { "$ref": "#/$defs/AuthenticationFlowObjectID" },
		"login_flow": { "$ref": "#/$defs/AuthenticationFlowObjectID" }
	}
}
`)

var _ = Schema.Add("AuthenticationFlowReauthFlow", `
{
	"type": "object",
	"required": ["id", "steps"],
	"properties": {
		"id": { "$ref": "#/$defs/AuthenticationFlowObjectID" },
		"steps": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/AuthenticationFlowReauthFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("AuthenticationFlowReauthFlowStep", `
{
	"type": "object",
	"required": ["type"],
	"properties": {
		"id": { "$ref": "#/$defs/AuthenticationFlowObjectID" },
		"type": {
			"type": "string",
			"enum": [
				"authenticate"
			]
		}
	},
	"allOf": [
		{
			"if": {
				"required": ["type"],
				"properties": {
					"type": { "const": "authenticate" }
				}
			},
			"then": {
				"required": ["one_of"],
				"properties": {
					"optional": { "type": "boolean" },
					"one_of": {
						"type": "array",
						"items": { "$ref": "#/$defs/AuthenticationFlowReauthFlowAuthenticate" }
					}
				}
			}
		}
	]
}
`)

var _ = Schema.Add("AuthenticationFlowReauthFlowAuthenticate", `
{
	"type": "object",
	"required": ["authentication"],
	"properties": {
		"authentication": {
			"type": "string",
			"enum": [
				"primary_password",
				"primary_passkey",
				"primary_oob_otp_email",
				"primary_oob_otp_sms",
				"secondary_password",
				"secondary_totp",
				"secondary_oob_otp_email",
				"secondary_oob_otp_sms"
			]
		},
		"target_step": { "$ref": "#/$defs/AuthenticationFlowObjectID" },
		"steps": {
			"type": "array",
			"items": { "$ref": "#/$defs/AuthenticationFlowReauthFlowStep" }
		}
	}
}
`)

type AuthenticationFlowObject interface {
	GetID() (string, bool)
	GetSteps() ([]AuthenticationFlowObject, bool)
	GetOneOf() ([]AuthenticationFlowObject, bool)
}

type AuthenticationFlowIdentification string

const (
	AuthenticationFlowIdentificationEmail    AuthenticationFlowIdentification = "email"
	AuthenticationFlowIdentificationPhone    AuthenticationFlowIdentification = "phone"
	AuthenticationFlowIdentificationUsername AuthenticationFlowIdentification = "username"
	AuthenticationFlowIdentificationOAuth    AuthenticationFlowIdentification = "oauth"
	AuthenticationFlowIdentificationPasskey  AuthenticationFlowIdentification = "passkey"
	AuthenticationFlowIdentificationSiwe     AuthenticationFlowIdentification = "siwe"
)

type AuthenticationFlowAuthentication string

const (
	AuthenticationFlowAuthenticationPrimaryPassword      AuthenticationFlowAuthentication = "primary_password"
	AuthenticationFlowAuthenticationPrimaryPasskey       AuthenticationFlowAuthentication = "primary_passkey"
	AuthenticationFlowAuthenticationPrimaryOOBOTPEmail   AuthenticationFlowAuthentication = "primary_oob_otp_email"
	AuthenticationFlowAuthenticationPrimaryOOBOTPSMS     AuthenticationFlowAuthentication = "primary_oob_otp_sms"
	AuthenticationFlowAuthenticationSecondaryPassword    AuthenticationFlowAuthentication = "secondary_password"
	AuthenticationFlowAuthenticationSecondaryTOTP        AuthenticationFlowAuthentication = "secondary_totp"
	AuthenticationFlowAuthenticationSecondaryOOBOTPEmail AuthenticationFlowAuthentication = "secondary_oob_otp_email"
	AuthenticationFlowAuthenticationSecondaryOOBOTPSMS   AuthenticationFlowAuthentication = "secondary_oob_otp_sms"
	AuthenticationFlowAuthenticationRecoveryCode         AuthenticationFlowAuthentication = "recovery_code"
	AuthenticationFlowAuthenticationDeviceToken          AuthenticationFlowAuthentication = "device_token"
)

func (m AuthenticationFlowAuthentication) AuthenticatorKind() model.AuthenticatorKind {
	switch m {
	case AuthenticationFlowAuthenticationPrimaryPassword:
		fallthrough
	case AuthenticationFlowAuthenticationPrimaryPasskey:
		fallthrough
	case AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
		fallthrough
	case AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
		return model.AuthenticatorKindPrimary
	case AuthenticationFlowAuthenticationSecondaryPassword:
		fallthrough
	case AuthenticationFlowAuthenticationSecondaryTOTP:
		fallthrough
	case AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
		fallthrough
	case AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
		return model.AuthenticatorKindSecondary
	case AuthenticationFlowAuthenticationRecoveryCode:
		panic(fmt.Errorf("%v is not an authenticator", m))
	case AuthenticationFlowAuthenticationDeviceToken:
		panic(fmt.Errorf("%v is not an authenticator", m))
	default:
		panic(fmt.Errorf("unknown authentication method: %v", m))
	}
}

type AuthenticationFlowConfig struct {
	SignupFlows      []*AuthenticationFlowSignupFlow      `json:"signup_flows,omitempty"`
	LoginFlows       []*AuthenticationFlowLoginFlow       `json:"login_flows,omitempty"`
	SignupLoginFlows []*AuthenticationFlowSignupLoginFlow `json:"signup_login_flows,omitempty"`
	ReauthFlows      []*AuthenticationFlowReauthFlow      `json:"reauth_flows,omitempty"`
}

type AuthenticationFlowSignupFlow struct {
	ID    string                              `json:"id,omitempty"`
	Steps []*AuthenticationFlowSignupFlowStep `json:"steps,omitempty"`
}

func (f *AuthenticationFlowSignupFlow) GetID() (string, bool) {
	return f.ID, true
}

func (f *AuthenticationFlowSignupFlow) GetSteps() ([]AuthenticationFlowObject, bool) {
	out := make([]AuthenticationFlowObject, len(f.Steps))
	for i, v := range f.Steps {
		v := v
		out[i] = v
	}
	return out, true
}

func (f *AuthenticationFlowSignupFlow) GetOneOf() ([]AuthenticationFlowObject, bool) {
	return nil, false
}

type AuthenticationFlowSignupFlowStepType string

const (
	AuthenticationFlowSignupFlowStepTypeIdentify     AuthenticationFlowSignupFlowStepType = "identify"
	AuthenticationFlowSignupFlowStepTypeAuthenticate AuthenticationFlowSignupFlowStepType = "authenticate"
	AuthenticationFlowSignupFlowStepTypeVerify       AuthenticationFlowSignupFlowStepType = "verify"
	AuthenticationFlowSignupFlowStepTypeUserProfile  AuthenticationFlowSignupFlowStepType = "user_profile"
	AuthenticationFlowSignupFlowStepTypeRecoveryCode AuthenticationFlowSignupFlowStepType = "recovery_code"
)

type AuthenticationFlowSignupFlowStep struct {
	ID   string                               `json:"id,omitempty"`
	Type AuthenticationFlowSignupFlowStepType `json:"type,omitempty"`

	// OneOf is relevant when Type is identify or authenticate.
	OneOf []*AuthenticationFlowSignupFlowOneOf `json:"one_of,omitempty"`
	// TargetStep is relevant when Type is verify.
	TargetStep string `json:"target_step,omitempty"`
	// UserProfile is relevant when Type is user_profile.
	UserProfile []*AuthenticationFlowSignupFlowUserProfile `json:"user_profile,omitempty"`
}

func (s *AuthenticationFlowSignupFlowStep) GetID() (string, bool) {
	return s.ID, true
}

func (s *AuthenticationFlowSignupFlowStep) GetSteps() ([]AuthenticationFlowObject, bool) {
	return nil, false
}

func (s *AuthenticationFlowSignupFlowStep) GetOneOf() ([]AuthenticationFlowObject, bool) {
	switch s.Type {
	case AuthenticationFlowSignupFlowStepTypeIdentify:
		fallthrough
	case AuthenticationFlowSignupFlowStepTypeAuthenticate:
		out := make([]AuthenticationFlowObject, len(s.OneOf))
		for i, v := range s.OneOf {
			v := v
			out[i] = v
		}
		return out, true
	default:
		return nil, false
	}
}

type AuthenticationFlowSignupFlowOneOf struct {
	// Identification is specific to identify.
	Identification AuthenticationFlowIdentification `json:"identification,omitempty"`

	// Authentication is specific to authenticate.
	Authentication AuthenticationFlowAuthentication `json:"authentication,omitempty"`
	// TargetStep is specific to authenticate.
	TargetStep string `json:"target_step,omitempty"`

	// Steps are common.
	Steps []*AuthenticationFlowSignupFlowStep `json:"steps,omitempty"`
}

func (f *AuthenticationFlowSignupFlowOneOf) GetID() (string, bool) {
	return "", false
}

func (f *AuthenticationFlowSignupFlowOneOf) GetSteps() ([]AuthenticationFlowObject, bool) {
	out := make([]AuthenticationFlowObject, len(f.Steps))
	for i, v := range f.Steps {
		v := v
		out[i] = v
	}
	return out, true
}

func (f *AuthenticationFlowSignupFlowOneOf) GetOneOf() ([]AuthenticationFlowObject, bool) {
	return nil, false
}

type AuthenticationFlowSignupFlowUserProfile struct {
	Pointer  string `json:"pointer,omitempty"`
	Required bool   `json:"required,omitempty"`
}

type AuthenticationFlowLoginFlow struct {
	ID    string                             `json:"id,omitempty"`
	Steps []*AuthenticationFlowLoginFlowStep `json:"steps,omitempty"`
}

func (f *AuthenticationFlowLoginFlow) GetID() (string, bool) {
	return f.ID, true
}

func (f *AuthenticationFlowLoginFlow) GetSteps() ([]AuthenticationFlowObject, bool) {
	out := make([]AuthenticationFlowObject, len(f.Steps))
	for i, v := range f.Steps {
		v := v
		out[i] = v
	}
	return out, true
}

func (f *AuthenticationFlowLoginFlow) GetOneOf() ([]AuthenticationFlowObject, bool) {
	return nil, false
}

type AuthenticationFlowLoginFlowStepType string

const (
	AuthenticationFlowLoginFlowStepTypeIdentify       AuthenticationFlowLoginFlowStepType = "identify"
	AuthenticationFlowLoginFlowStepTypeAuthenticate   AuthenticationFlowLoginFlowStepType = "authenticate"
	AuthenticationFlowLoginFlowStepTypeChangePassword AuthenticationFlowLoginFlowStepType = "change_password"
)

type AuthenticationFlowLoginFlowStep struct {
	ID   string                              `json:"id,omitempty"`
	Type AuthenticationFlowLoginFlowStepType `json:"type,omitempty"`

	// Optional is relevant when Type is authenticate.
	Optional *bool `json:"optional,omitempty"`

	// OneOf is relevant when Type is identify or authenticate.
	OneOf []*AuthenticationFlowLoginFlowOneOf `json:"one_of,omitempty"`

	// TargetStep is relevant when Type is change_password.
	TargetStep string `json:"target_step,omitempty"`
}

func (s *AuthenticationFlowLoginFlowStep) GetID() (string, bool) {
	return s.ID, true
}

func (s *AuthenticationFlowLoginFlowStep) GetSteps() ([]AuthenticationFlowObject, bool) {
	return nil, false
}

func (s *AuthenticationFlowLoginFlowStep) GetOneOf() ([]AuthenticationFlowObject, bool) {
	switch s.Type {
	case AuthenticationFlowLoginFlowStepTypeIdentify:
		fallthrough
	case AuthenticationFlowLoginFlowStepTypeAuthenticate:
		out := make([]AuthenticationFlowObject, len(s.OneOf))
		for i, v := range s.OneOf {
			v := v
			out[i] = v
		}
		return out, true
	default:
		return nil, false
	}
}

type AuthenticationFlowLoginFlowOneOf struct {
	// Identification is specific to identify.
	Identification AuthenticationFlowIdentification `json:"identification,omitempty"`

	// Authentication is specific to authenticate.
	Authentication AuthenticationFlowAuthentication `json:"authentication,omitempty"`
	// TargetStep is specific to authenticate.
	TargetStep string `json:"target_step,omitempty"`

	// Steps are common.
	Steps []*AuthenticationFlowLoginFlowStep `json:"steps,omitempty"`
}

func (f *AuthenticationFlowLoginFlowOneOf) GetID() (string, bool) {
	return "", false
}

func (f *AuthenticationFlowLoginFlowOneOf) GetSteps() ([]AuthenticationFlowObject, bool) {
	out := make([]AuthenticationFlowObject, len(f.Steps))
	for i, v := range f.Steps {
		v := v
		out[i] = v
	}
	return out, true
}

func (f *AuthenticationFlowLoginFlowOneOf) GetOneOf() ([]AuthenticationFlowObject, bool) {
	return nil, false
}

type AuthenticationFlowSignupLoginFlow struct {
	ID    string                                   `json:"id,omitempty"`
	Steps []*AuthenticationFlowSignupLoginFlowStep `json:"steps,omitempty"`
}

func (f *AuthenticationFlowSignupLoginFlow) GetID() (string, bool) {
	return f.ID, true
}

func (f *AuthenticationFlowSignupLoginFlow) GetSteps() ([]AuthenticationFlowObject, bool) {
	out := make([]AuthenticationFlowObject, len(f.Steps))
	for i, v := range f.Steps {
		v := v
		out[i] = v
	}
	return out, true
}

func (f *AuthenticationFlowSignupLoginFlow) GetOneOf() ([]AuthenticationFlowObject, bool) {
	return nil, false
}

type AuthenticationFlowSignupLoginFlowStep struct {
	ID    string                                    `json:"id,omitempty"`
	Type  AuthenticationFlowSignupLoginFlowStepType `json:"type,omitempty"`
	OneOf []*AuthenticationFlowSignupLoginFlowOneOf `json:"one_of,omitempty"`
}

func (s *AuthenticationFlowSignupLoginFlowStep) GetID() (string, bool) {
	return s.ID, true
}

func (s *AuthenticationFlowSignupLoginFlowStep) GetSteps() ([]AuthenticationFlowObject, bool) {
	return nil, false
}

func (s *AuthenticationFlowSignupLoginFlowStep) GetOneOf() ([]AuthenticationFlowObject, bool) {
	switch s.Type {
	case AuthenticationFlowSignupLoginFlowStepTypeIdentify:
		out := make([]AuthenticationFlowObject, len(s.OneOf))
		for i, v := range s.OneOf {
			v := v
			out[i] = v
		}
		return out, true
	default:
		return nil, false
	}
}

type AuthenticationFlowSignupLoginFlowStepType string

const (
	AuthenticationFlowSignupLoginFlowStepTypeIdentify AuthenticationFlowSignupLoginFlowStepType = "identify"
)

type AuthenticationFlowSignupLoginFlowOneOf struct {
	Identification AuthenticationFlowIdentification `json:"identification,omitempty"`
	SignupFlow     string                           `json:"signup_flow,omitempty"`
	LoginFlow      string                           `json:"login_flow,omitempty"`
}

func (s *AuthenticationFlowSignupLoginFlowOneOf) GetID() (string, bool) {
	return "", false
}

func (s *AuthenticationFlowSignupLoginFlowOneOf) GetSteps() ([]AuthenticationFlowObject, bool) {
	return nil, false
}

func (s *AuthenticationFlowSignupLoginFlowOneOf) GetOneOf() ([]AuthenticationFlowObject, bool) {
	return nil, false
}

type AuthenticationFlowReauthFlow struct {
	ID    string                              `json:"id,omitempty"`
	Steps []*AuthenticationFlowReauthFlowStep `json:"steps,omitempty"`
}

func (f *AuthenticationFlowReauthFlow) GetID() (string, bool) {
	return f.ID, true
}

func (f *AuthenticationFlowReauthFlow) GetSteps() ([]AuthenticationFlowObject, bool) {
	out := make([]AuthenticationFlowObject, len(f.Steps))
	for i, v := range f.Steps {
		v := v
		out[i] = v
	}
	return out, true
}

func (f *AuthenticationFlowReauthFlow) GetOneOf() ([]AuthenticationFlowObject, bool) {
	return nil, false
}

type AuthenticationFlowReauthFlowStepType string

const (
	AuthenticationFlowReauthFlowStepTypeAuthenticate AuthenticationFlowReauthFlowStepType = "authenticate"
)

type AuthenticationFlowReauthFlowStep struct {
	ID   string                               `json:"id,omitempty"`
	Type AuthenticationFlowReauthFlowStepType `json:"type,omitempty"`

	// Optional is relevant when Type is authenticate.
	Optional *bool `json:"optional,omitempty"`

	// OneOf is relevant when Type is authenticate.
	OneOf []*AuthenticationFlowReauthFlowOneOf `json:"one_of,omitempty"`
}

func (s *AuthenticationFlowReauthFlowStep) GetID() (string, bool) {
	return s.ID, true
}

func (s *AuthenticationFlowReauthFlowStep) GetSteps() ([]AuthenticationFlowObject, bool) {
	return nil, false
}

func (s *AuthenticationFlowReauthFlowStep) GetOneOf() ([]AuthenticationFlowObject, bool) {
	switch s.Type {
	case AuthenticationFlowReauthFlowStepTypeAuthenticate:
		out := make([]AuthenticationFlowObject, len(s.OneOf))
		for i, v := range s.OneOf {
			v := v
			out[i] = v
		}
		return out, true
	default:
		return nil, false
	}
}

type AuthenticationFlowReauthFlowOneOf struct {
	Authentication AuthenticationFlowAuthentication    `json:"authentication,omitempty"`
	TargetStep     string                              `json:"target_step,omitempty"`
	Steps          []*AuthenticationFlowReauthFlowStep `json:"steps,omitempty"`
}

func (f *AuthenticationFlowReauthFlowOneOf) GetID() (string, bool) {
	return "", false
}

func (f *AuthenticationFlowReauthFlowOneOf) GetSteps() ([]AuthenticationFlowObject, bool) {
	out := make([]AuthenticationFlowObject, len(f.Steps))
	for i, v := range f.Steps {
		v := v
		out[i] = v
	}
	return out, true
}

func (f *AuthenticationFlowReauthFlowOneOf) GetOneOf() ([]AuthenticationFlowObject, bool) {
	return nil, false
}
