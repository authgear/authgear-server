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

var _ = Schema.Add("AuthenticationFlowIdentification", `
{
	"type": "string",
	"enum": [
		"email",
		"phone",
		"username",
		"oauth",
		"passkey"
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
				"recovery_code",
				"prompt_create_passkey"
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
		"identification": { "$ref": "#/$defs/AuthenticationFlowIdentification" },
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
				"change_password",
				"prompt_create_passkey"
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
		"identification": { "$ref": "#/$defs/AuthenticationFlowIdentification" },
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
		"identification": { "$ref": "#/$defs/AuthenticationFlowIdentification" },
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
	IsFlowObject()
}

type AuthenticationFlowObjectFlowRoot interface {
	AuthenticationFlowObject
	GetID() string
	GetSteps() []AuthenticationFlowObject
}

type AuthenticationFlowObjectFlowStep interface {
	AuthenticationFlowObject
	GetID() string
	GetType() string
	GetOneOf() []AuthenticationFlowObject
}

type AuthenticationFlowObjectFlowBranchInfo struct {
	Identification AuthenticationFlowIdentification `json:"identification,omitempty"`
	Authentication AuthenticationFlowAuthentication `json:"authentication,omitempty"`
}

type AuthenticationFlowObjectFlowBranch interface {
	AuthenticationFlowObject
	GetSteps() []AuthenticationFlowObject
	GetBranchInfo() AuthenticationFlowObjectFlowBranchInfo
}

type AuthenticationFlowIdentification string

const (
	AuthenticationFlowIdentificationEmail    AuthenticationFlowIdentification = "email"
	AuthenticationFlowIdentificationPhone    AuthenticationFlowIdentification = "phone"
	AuthenticationFlowIdentificationUsername AuthenticationFlowIdentification = "username"
	AuthenticationFlowIdentificationOAuth    AuthenticationFlowIdentification = "oauth"
	AuthenticationFlowIdentificationPasskey  AuthenticationFlowIdentification = "passkey"
)

func (m AuthenticationFlowIdentification) PrimaryAuthentications() []AuthenticationFlowAuthentication {
	switch m {
	case AuthenticationFlowIdentificationEmail:
		return []AuthenticationFlowAuthentication{
			AuthenticationFlowAuthenticationPrimaryPassword,
			AuthenticationFlowAuthenticationPrimaryOOBOTPEmail,
			AuthenticationFlowAuthenticationPrimaryPasskey,
		}
	case AuthenticationFlowIdentificationPhone:
		return []AuthenticationFlowAuthentication{
			AuthenticationFlowAuthenticationPrimaryPassword,
			AuthenticationFlowAuthenticationPrimaryOOBOTPSMS,
			AuthenticationFlowAuthenticationPrimaryPasskey,
		}
	case AuthenticationFlowIdentificationUsername:
		return []AuthenticationFlowAuthentication{
			AuthenticationFlowAuthenticationPrimaryPassword,
			AuthenticationFlowAuthenticationPrimaryPasskey,
		}
	case AuthenticationFlowIdentificationOAuth:
		// OAuth does not require primary authentication.
		return nil
	case AuthenticationFlowIdentificationPasskey:
		// Passkey does not require primary authentication.
		return nil
	default:
		panic(fmt.Errorf("unknown identification: %v", m))
	}
}

func (m AuthenticationFlowIdentification) SecondaryAuthentications() []AuthenticationFlowAuthentication {
	all := []AuthenticationFlowAuthentication{
		AuthenticationFlowAuthenticationSecondaryPassword,
		AuthenticationFlowAuthenticationSecondaryTOTP,
		AuthenticationFlowAuthenticationSecondaryOOBOTPEmail,
		AuthenticationFlowAuthenticationSecondaryOOBOTPSMS,
	}
	switch m {
	case AuthenticationFlowIdentificationEmail:
		return all
	case AuthenticationFlowIdentificationPhone:
		return all
	case AuthenticationFlowIdentificationUsername:
		return all
	case AuthenticationFlowIdentificationOAuth:
		// OAuth does not require secondary authentication.
		return nil
	case AuthenticationFlowIdentificationPasskey:
		// Passkey does not require secondary authentication.
		return nil
	default:
		panic(fmt.Errorf("unknown identification: %v", m))
	}
}

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
		panic(fmt.Errorf("unknown authentication: %v", m))
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

var _ AuthenticationFlowObjectFlowRoot = &AuthenticationFlowSignupFlow{}

func (f *AuthenticationFlowSignupFlow) IsFlowObject() {}
func (f *AuthenticationFlowSignupFlow) GetID() string { return f.ID }
func (f *AuthenticationFlowSignupFlow) GetSteps() []AuthenticationFlowObject {
	out := make([]AuthenticationFlowObject, len(f.Steps))
	for i, v := range f.Steps {
		v := v
		out[i] = v
	}
	return out
}

type AuthenticationFlowSignupFlowStepType string

const (
	AuthenticationFlowSignupFlowStepTypeIdentify            AuthenticationFlowSignupFlowStepType = "identify"
	AuthenticationFlowSignupFlowStepTypeAuthenticate        AuthenticationFlowSignupFlowStepType = "authenticate"
	AuthenticationFlowSignupFlowStepTypeVerify              AuthenticationFlowSignupFlowStepType = "verify"
	AuthenticationFlowSignupFlowStepTypeUserProfile         AuthenticationFlowSignupFlowStepType = "user_profile"
	AuthenticationFlowSignupFlowStepTypeRecoveryCode        AuthenticationFlowSignupFlowStepType = "recovery_code"
	AuthenticationFlowSignupFlowStepTypePromptCreatePasskey AuthenticationFlowSignupFlowStepType = "prompt_create_passkey"
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

var _ AuthenticationFlowObjectFlowStep = &AuthenticationFlowSignupFlowStep{}

func (s *AuthenticationFlowSignupFlowStep) IsFlowObject()   {}
func (s *AuthenticationFlowSignupFlowStep) GetID() string   { return s.ID }
func (s *AuthenticationFlowSignupFlowStep) GetType() string { return string(s.Type) }
func (s *AuthenticationFlowSignupFlowStep) GetOneOf() []AuthenticationFlowObject {
	switch s.Type {
	case AuthenticationFlowSignupFlowStepTypeIdentify:
		fallthrough
	case AuthenticationFlowSignupFlowStepTypeAuthenticate:
		out := make([]AuthenticationFlowObject, len(s.OneOf))
		for i, v := range s.OneOf {
			v := v
			out[i] = v
		}
		return out
	default:
		return nil
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

var _ AuthenticationFlowObjectFlowBranch = &AuthenticationFlowSignupFlowOneOf{}

func (f *AuthenticationFlowSignupFlowOneOf) IsFlowObject() {}

func (f *AuthenticationFlowSignupFlowOneOf) GetSteps() []AuthenticationFlowObject {
	out := make([]AuthenticationFlowObject, len(f.Steps))
	for i, v := range f.Steps {
		v := v
		out[i] = v
	}
	return out
}

func (f *AuthenticationFlowSignupFlowOneOf) GetBranchInfo() AuthenticationFlowObjectFlowBranchInfo {
	return AuthenticationFlowObjectFlowBranchInfo{
		Identification: f.Identification,
		Authentication: f.Authentication,
	}
}

type AuthenticationFlowSignupFlowUserProfile struct {
	Pointer  string `json:"pointer,omitempty"`
	Required bool   `json:"required,omitempty"`
}

type AuthenticationFlowLoginFlow struct {
	ID    string                             `json:"id,omitempty"`
	Steps []*AuthenticationFlowLoginFlowStep `json:"steps,omitempty"`
}

var _ AuthenticationFlowObjectFlowRoot = &AuthenticationFlowLoginFlow{}

func (f *AuthenticationFlowLoginFlow) IsFlowObject() {}

func (f *AuthenticationFlowLoginFlow) GetID() string {
	return f.ID
}

func (f *AuthenticationFlowLoginFlow) GetSteps() []AuthenticationFlowObject {
	out := make([]AuthenticationFlowObject, len(f.Steps))
	for i, v := range f.Steps {
		v := v
		out[i] = v
	}
	return out
}

type AuthenticationFlowLoginFlowStepType string

const (
	AuthenticationFlowLoginFlowStepTypeIdentify            AuthenticationFlowLoginFlowStepType = "identify"
	AuthenticationFlowLoginFlowStepTypeAuthenticate        AuthenticationFlowLoginFlowStepType = "authenticate"
	AuthenticationFlowLoginFlowStepTypeChangePassword      AuthenticationFlowLoginFlowStepType = "change_password"
	AuthenticationFlowLoginFlowStepTypePromptCreatePasskey AuthenticationFlowLoginFlowStepType = "prompt_create_passkey"
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

var _ AuthenticationFlowObjectFlowStep = &AuthenticationFlowLoginFlowStep{}

func (s *AuthenticationFlowLoginFlowStep) IsFlowObject()   {}
func (s *AuthenticationFlowLoginFlowStep) GetID() string   { return s.ID }
func (s *AuthenticationFlowLoginFlowStep) GetType() string { return string(s.Type) }

func (s *AuthenticationFlowLoginFlowStep) GetOneOf() []AuthenticationFlowObject {
	switch s.Type {
	case AuthenticationFlowLoginFlowStepTypeIdentify:
		fallthrough
	case AuthenticationFlowLoginFlowStepTypeAuthenticate:
		out := make([]AuthenticationFlowObject, len(s.OneOf))
		for i, v := range s.OneOf {
			v := v
			out[i] = v
		}
		return out
	default:
		return nil
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

var _ AuthenticationFlowObjectFlowBranch = &AuthenticationFlowLoginFlowOneOf{}

func (f *AuthenticationFlowLoginFlowOneOf) IsFlowObject() {}

func (f *AuthenticationFlowLoginFlowOneOf) GetSteps() []AuthenticationFlowObject {
	out := make([]AuthenticationFlowObject, len(f.Steps))
	for i, v := range f.Steps {
		v := v
		out[i] = v
	}
	return out
}

func (f *AuthenticationFlowLoginFlowOneOf) GetBranchInfo() AuthenticationFlowObjectFlowBranchInfo {
	return AuthenticationFlowObjectFlowBranchInfo{
		Identification: f.Identification,
		Authentication: f.Authentication,
	}
}

type AuthenticationFlowSignupLoginFlow struct {
	ID    string                                   `json:"id,omitempty"`
	Steps []*AuthenticationFlowSignupLoginFlowStep `json:"steps,omitempty"`
}

var _ AuthenticationFlowObjectFlowRoot = &AuthenticationFlowSignupLoginFlow{}

func (f *AuthenticationFlowSignupLoginFlow) IsFlowObject() {}
func (f *AuthenticationFlowSignupLoginFlow) GetID() string { return f.ID }

func (f *AuthenticationFlowSignupLoginFlow) GetSteps() []AuthenticationFlowObject {
	out := make([]AuthenticationFlowObject, len(f.Steps))
	for i, v := range f.Steps {
		v := v
		out[i] = v
	}
	return out
}

type AuthenticationFlowSignupLoginFlowStep struct {
	ID    string                                    `json:"id,omitempty"`
	Type  AuthenticationFlowSignupLoginFlowStepType `json:"type,omitempty"`
	OneOf []*AuthenticationFlowSignupLoginFlowOneOf `json:"one_of,omitempty"`
}

var _ AuthenticationFlowObjectFlowStep = &AuthenticationFlowSignupLoginFlowStep{}

func (s *AuthenticationFlowSignupLoginFlowStep) IsFlowObject()   {}
func (s *AuthenticationFlowSignupLoginFlowStep) GetID() string   { return s.ID }
func (s *AuthenticationFlowSignupLoginFlowStep) GetType() string { return string(s.Type) }

func (s *AuthenticationFlowSignupLoginFlowStep) GetOneOf() []AuthenticationFlowObject {
	switch s.Type {
	case AuthenticationFlowSignupLoginFlowStepTypeIdentify:
		out := make([]AuthenticationFlowObject, len(s.OneOf))
		for i, v := range s.OneOf {
			v := v
			out[i] = v
		}
		return out
	default:
		return nil
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

var _ AuthenticationFlowObjectFlowBranch = &AuthenticationFlowSignupLoginFlowOneOf{}

func (s *AuthenticationFlowSignupLoginFlowOneOf) IsFlowObject() {}

func (s *AuthenticationFlowSignupLoginFlowOneOf) GetSteps() []AuthenticationFlowObject {
	return nil
}

func (s *AuthenticationFlowSignupLoginFlowOneOf) GetBranchInfo() AuthenticationFlowObjectFlowBranchInfo {
	return AuthenticationFlowObjectFlowBranchInfo{
		Identification: s.Identification,
	}
}

type AuthenticationFlowReauthFlow struct {
	ID    string                              `json:"id,omitempty"`
	Steps []*AuthenticationFlowReauthFlowStep `json:"steps,omitempty"`
}

var _ AuthenticationFlowObjectFlowRoot = &AuthenticationFlowReauthFlow{}

func (f *AuthenticationFlowReauthFlow) IsFlowObject() {}
func (f *AuthenticationFlowReauthFlow) GetID() string { return f.ID }

func (f *AuthenticationFlowReauthFlow) GetSteps() []AuthenticationFlowObject {
	out := make([]AuthenticationFlowObject, len(f.Steps))
	for i, v := range f.Steps {
		v := v
		out[i] = v
	}
	return out
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

var _ AuthenticationFlowObjectFlowStep = &AuthenticationFlowReauthFlowStep{}

func (s *AuthenticationFlowReauthFlowStep) IsFlowObject()   {}
func (s *AuthenticationFlowReauthFlowStep) GetID() string   { return s.ID }
func (s *AuthenticationFlowReauthFlowStep) GetType() string { return string(s.Type) }

func (s *AuthenticationFlowReauthFlowStep) GetOneOf() []AuthenticationFlowObject {
	switch s.Type {
	case AuthenticationFlowReauthFlowStepTypeAuthenticate:
		out := make([]AuthenticationFlowObject, len(s.OneOf))
		for i, v := range s.OneOf {
			v := v
			out[i] = v
		}
		return out
	default:
		return nil
	}
}

type AuthenticationFlowReauthFlowOneOf struct {
	Authentication AuthenticationFlowAuthentication    `json:"authentication,omitempty"`
	TargetStep     string                              `json:"target_step,omitempty"`
	Steps          []*AuthenticationFlowReauthFlowStep `json:"steps,omitempty"`
}

var _ AuthenticationFlowObjectFlowBranch = &AuthenticationFlowReauthFlowOneOf{}

func (f *AuthenticationFlowReauthFlowOneOf) IsFlowObject() {}

func (f *AuthenticationFlowReauthFlowOneOf) GetSteps() []AuthenticationFlowObject {
	out := make([]AuthenticationFlowObject, len(f.Steps))
	for i, v := range f.Steps {
		v := v
		out[i] = v
	}
	return out
}

func (f *AuthenticationFlowReauthFlowOneOf) GetBranchInfo() AuthenticationFlowObjectFlowBranchInfo {
	return AuthenticationFlowObjectFlowBranchInfo{
		Authentication: f.Authentication,
	}
}
