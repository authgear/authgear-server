package config

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
)

var _ = Schema.Add("WorkflowConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"signup_flows": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/WorkflowSignupFlow" }
		},
		"login_flows": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/WorkflowLoginFlow" }
		},
		"signup_login_flows": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/WorkflowSignupLoginFlow" }
		},
		"reauth_flows": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/WorkflowReauthFlow" }
		}
	}
}
`)

var _ = Schema.Add("WorkflowObjectID", `
{
	"type": "string",
	"pattern": "^[a-zA-Z_][a-zA-Z0-9_]*$"
}
`)

var _ = Schema.Add("WorkflowIdentificationMethod", `
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

var _ = Schema.Add("WorkflowSignupFlow", `
{
	"type": "object",
	"required": ["id", "steps"],
	"properties": {
		"id": { "$ref": "#/$defs/WorkflowObjectID" },
		"steps": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/WorkflowSignupFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("WorkflowSignupFlowStep", `
{
	"type": "object",
	"required": ["type"],
	"properties": {
		"id": { "$ref": "#/$defs/WorkflowObjectID" },
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
						"items": { "$ref": "#/$defs/WorkflowSignupFlowIdentify" }
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
						"items": { "$ref": "#/$defs/WorkflowSignupFlowAuthenticate" }
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
					"target_step": { "$ref": "#/$defs/WorkflowObjectID" }
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
						"items": { "$ref": "#/$defs/WorkflowSignupFlowUserProfile" }
					}
				}
			}
		}
	]
}
`)

var _ = Schema.Add("WorkflowSignupFlowIdentify", `
{
	"type": "object",
	"required": ["identification"],
	"properties": {
		"identification": { "$ref": "#/$defs/WorkflowIdentificationMethod" },
		"steps": {
			"type": "array",
			"items": { "$ref": "#/$defs/WorkflowSignupFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("WorkflowSignupFlowAuthenticate", `
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
		"target_step": { "$ref": "#/$defs/WorkflowObjectID" },
		"steps": {
			"type": "array",
			"items": { "$ref": "#/$defs/WorkflowSignupFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("WorkflowSignupFlowUserProfile", `
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

var _ = Schema.Add("WorkflowLoginFlow", `
{
	"type": "object",
	"required": ["id", "steps"],
	"properties": {
		"id": { "$ref": "#/$defs/WorkflowObjectID" },
		"steps": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/WorkflowLoginFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("WorkflowLoginFlowStep", `
{
	"type": "object",
	"required": ["type"],
	"properties": {
		"id": { "$ref": "#/$defs/WorkflowObjectID" },
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
						"items": { "$ref": "#/$defs/WorkflowLoginFlowIdentify" }
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
						"items": { "$ref": "#/$defs/WorkflowLoginFlowAuthenticate" }
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
					"target_step": { "$ref": "#/$defs/WorkflowObjectID" }
				}
			}
		}
	]
}
`)

var _ = Schema.Add("WorkflowLoginFlowIdentify", `
{
	"type": "object",
	"required": ["identification"],
	"properties": {
		"identification": { "$ref": "#/$defs/WorkflowIdentificationMethod" },
		"steps": {
			"type": "array",
			"items": { "$ref": "#/$defs/WorkflowLoginFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("WorkflowLoginFlowAuthenticate", `
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
		"target_step": { "$ref": "#/$defs/WorkflowObjectID" },
		"steps": {
			"type": "array",
			"items": { "$ref": "#/$defs/WorkflowLoginFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("WorkflowSignupLoginFlow", `
{
	"type": "object",
	"required": ["id", "steps"],
	"properties": {
		"id": { "$ref": "#/$defs/WorkflowObjectID" },
		"steps": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/WorkflowSignupLoginFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("WorkflowSignupLoginFlowStep", `
{
	"type": "object",
	"required": ["type"],
	"properties": {
		"id": { "$ref": "#/$defs/WorkflowObjectID" },
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
						"items": { "$ref": "#/$defs/WorkflowSignupLoginFlowIdentify" }
					}
				}
			}
		}
	]
}
`)

var _ = Schema.Add("WorkflowSignupLoginFlowIdentify", `
{
	"type": "object",
	"required": ["identification", "signup_flow", "login_flow"],
	"properties": {
		"identification": { "$ref": "#/$defs/WorkflowIdentificationMethod" },
		"signup_flow": { "$ref": "#/$defs/WorkflowObjectID" },
		"login_flow": { "$ref": "#/$defs/WorkflowObjectID" }
	}
}
`)

var _ = Schema.Add("WorkflowReauthFlow", `
{
	"type": "object",
	"required": ["id", "steps"],
	"properties": {
		"id": { "$ref": "#/$defs/WorkflowObjectID" },
		"steps": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/WorkflowReauthFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("WorkflowReauthFlowStep", `
{
	"type": "object",
	"required": ["type"],
	"properties": {
		"id": { "$ref": "#/$defs/WorkflowObjectID" },
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
						"items": { "$ref": "#/$defs/WorkflowReauthFlowAuthenticate" }
					}
				}
			}
		}
	]
}
`)

var _ = Schema.Add("WorkflowReauthFlowAuthenticate", `
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
		"target_step": { "$ref": "#/$defs/WorkflowObjectID" },
		"steps": {
			"type": "array",
			"items": { "$ref": "#/$defs/WorkflowReauthFlowStep" }
		}
	}
}
`)

type WorkflowObject interface {
	GetID() (string, bool)
	GetSteps() ([]WorkflowObject, bool)
	GetOneOf() ([]WorkflowObject, bool)
}

type WorkflowIdentificationMethod string

const (
	WorkflowIdentificationMethodEmail    WorkflowIdentificationMethod = "email"
	WorkflowIdentificationMethodPhone    WorkflowIdentificationMethod = "phone"
	WorkflowIdentificationMethodUsername WorkflowIdentificationMethod = "username"
	WorkflowIdentificationMethodOAuth    WorkflowIdentificationMethod = "oauth"
	WorkflowIdentificationMethodPasskey  WorkflowIdentificationMethod = "passkey"
	WorkflowIdentificationMethodSiwe     WorkflowIdentificationMethod = "siwe"
)

type WorkflowAuthenticationMethod string

const (
	WorkflowAuthenticationMethodPrimaryPassword      WorkflowAuthenticationMethod = "primary_password"
	WorkflowAuthenticationMethodPrimaryPasskey       WorkflowAuthenticationMethod = "primary_passkey"
	WorkflowAuthenticationMethodPrimaryOOBOTPEmail   WorkflowAuthenticationMethod = "primary_oob_otp_email"
	WorkflowAuthenticationMethodPrimaryOOBOTPSMS     WorkflowAuthenticationMethod = "primary_oob_otp_sms"
	WorkflowAuthenticationMethodSecondaryPassword    WorkflowAuthenticationMethod = "secondary_password"
	WorkflowAuthenticationMethodSecondaryTOTP        WorkflowAuthenticationMethod = "secondary_totp"
	WorkflowAuthenticationMethodSecondaryOOBOTPEmail WorkflowAuthenticationMethod = "secondary_oob_otp_email"
	WorkflowAuthenticationMethodSecondaryOOBOTPSMS   WorkflowAuthenticationMethod = "secondary_oob_otp_sms"
	WorkflowAuthenticationMethodRecoveryCode         WorkflowAuthenticationMethod = "recovery_code"
	WorkflowAuthenticationMethodDeviceToken          WorkflowAuthenticationMethod = "device_token"
)

func (m WorkflowAuthenticationMethod) AuthenticatorKind() model.AuthenticatorKind {
	switch m {
	case WorkflowAuthenticationMethodPrimaryPassword:
		fallthrough
	case WorkflowAuthenticationMethodPrimaryPasskey:
		fallthrough
	case WorkflowAuthenticationMethodPrimaryOOBOTPEmail:
		fallthrough
	case WorkflowAuthenticationMethodPrimaryOOBOTPSMS:
		return model.AuthenticatorKindPrimary
	case WorkflowAuthenticationMethodSecondaryPassword:
		fallthrough
	case WorkflowAuthenticationMethodSecondaryTOTP:
		fallthrough
	case WorkflowAuthenticationMethodSecondaryOOBOTPEmail:
		fallthrough
	case WorkflowAuthenticationMethodSecondaryOOBOTPSMS:
		return model.AuthenticatorKindSecondary
	case WorkflowAuthenticationMethodRecoveryCode:
		panic(fmt.Errorf("%v is not an authenticator", m))
	case WorkflowAuthenticationMethodDeviceToken:
		panic(fmt.Errorf("%v is not an authenticator", m))
	default:
		panic(fmt.Errorf("unknown authentication method: %v", m))
	}
}

type WorkflowConfig struct {
	SignupFlows      []*WorkflowSignupFlow      `json:"signup_flows,omitempty"`
	LoginFlows       []*WorkflowLoginFlow       `json:"login_flows,omitempty"`
	SignupLoginFlows []*WorkflowSignupLoginFlow `json:"signup_login_flows,omitempty"`
	ReauthFlows      []*WorkflowReauthFlow      `json:"reauth_flows,omitempty"`
}

type WorkflowSignupFlow struct {
	ID    string                    `json:"id,omitempty"`
	Steps []*WorkflowSignupFlowStep `json:"steps,omitempty"`
}

func (f *WorkflowSignupFlow) GetID() (string, bool) {
	return f.ID, true
}

func (f *WorkflowSignupFlow) GetSteps() ([]WorkflowObject, bool) {
	out := make([]WorkflowObject, len(f.Steps))
	for i, v := range f.Steps {
		v := v
		out[i] = v
	}
	return out, true
}

func (f *WorkflowSignupFlow) GetOneOf() ([]WorkflowObject, bool) {
	return nil, false
}

type WorkflowSignupFlowStepType string

const (
	WorkflowSignupFlowStepTypeIdentify     WorkflowSignupFlowStepType = "identify"
	WorkflowSignupFlowStepTypeAuthenticate WorkflowSignupFlowStepType = "authenticate"
	WorkflowSignupFlowStepTypeVerify       WorkflowSignupFlowStepType = "verify"
	WorkflowSignupFlowStepTypeUserProfile  WorkflowSignupFlowStepType = "user_profile"
	WorkflowSignupFlowStepTypeRecoveryCode WorkflowSignupFlowStepType = "recovery_code"
)

type WorkflowSignupFlowStep struct {
	ID   string                     `json:"id,omitempty"`
	Type WorkflowSignupFlowStepType `json:"type,omitempty"`

	// OneOf is relevant when Type is identify or authenticate.
	OneOf []*WorkflowSignupFlowOneOf `json:"one_of,omitempty"`
	// TargetStep is relevant when Type is verify.
	TargetStep string `json:"target_step,omitempty"`
	// UserProfile is relevant when Type is user_profile.
	UserProfile []*WorkflowSignupFlowUserProfile `json:"user_profile,omitempty"`
}

func (s *WorkflowSignupFlowStep) GetID() (string, bool) {
	return s.ID, true
}

func (s *WorkflowSignupFlowStep) GetSteps() ([]WorkflowObject, bool) {
	return nil, false
}

func (s *WorkflowSignupFlowStep) GetOneOf() ([]WorkflowObject, bool) {
	switch s.Type {
	case WorkflowSignupFlowStepTypeIdentify:
		fallthrough
	case WorkflowSignupFlowStepTypeAuthenticate:
		out := make([]WorkflowObject, len(s.OneOf))
		for i, v := range s.OneOf {
			v := v
			out[i] = v
		}
		return out, true
	default:
		return nil, false
	}
}

type WorkflowSignupFlowOneOf struct {
	// Identification is specific to identify.
	Identification WorkflowIdentificationMethod `json:"identification,omitempty"`

	// Authentication is specific to authenticate.
	Authentication WorkflowAuthenticationMethod `json:"authentication,omitempty"`
	// TargetStep is specific to authenticate.
	TargetStep string `json:"target_step,omitempty"`

	// Steps are common.
	Steps []*WorkflowSignupFlowStep `json:"steps,omitempty"`
}

func (f *WorkflowSignupFlowOneOf) GetID() (string, bool) {
	return "", false
}

func (f *WorkflowSignupFlowOneOf) GetSteps() ([]WorkflowObject, bool) {
	out := make([]WorkflowObject, len(f.Steps))
	for i, v := range f.Steps {
		v := v
		out[i] = v
	}
	return out, true
}

func (f *WorkflowSignupFlowOneOf) GetOneOf() ([]WorkflowObject, bool) {
	return nil, false
}

type WorkflowSignupFlowUserProfile struct {
	Pointer  string `json:"pointer,omitempty"`
	Required bool   `json:"required,omitempty"`
}

type WorkflowLoginFlow struct {
	ID    string                   `json:"id,omitempty"`
	Steps []*WorkflowLoginFlowStep `json:"steps,omitempty"`
}

func (f *WorkflowLoginFlow) GetID() (string, bool) {
	return f.ID, true
}

func (f *WorkflowLoginFlow) GetSteps() ([]WorkflowObject, bool) {
	out := make([]WorkflowObject, len(f.Steps))
	for i, v := range f.Steps {
		v := v
		out[i] = v
	}
	return out, true
}

func (f *WorkflowLoginFlow) GetOneOf() ([]WorkflowObject, bool) {
	return nil, false
}

type WorkflowLoginFlowStepType string

const (
	WorkflowLoginFlowStepTypeIdentify       WorkflowLoginFlowStepType = "identify"
	WorkflowLoginFlowStepTypeAuthenticate   WorkflowLoginFlowStepType = "authenticate"
	WorkflowLoginFlowStepTypeChangePassword WorkflowLoginFlowStepType = "change_password"
)

type WorkflowLoginFlowStep struct {
	ID   string                    `json:"id,omitempty"`
	Type WorkflowLoginFlowStepType `json:"type,omitempty"`

	// Optional is relevant when Type is authenticate.
	Optional *bool `json:"optional,omitempty"`

	// OneOf is relevant when Type is identify or authenticate.
	OneOf []*WorkflowLoginFlowOneOf `json:"one_of,omitempty"`

	// TargetStep is relevant when Type is change_password.
	TargetStep string `json:"target_step,omitempty"`
}

func (s *WorkflowLoginFlowStep) GetID() (string, bool) {
	return s.ID, true
}

func (s *WorkflowLoginFlowStep) GetSteps() ([]WorkflowObject, bool) {
	return nil, false
}

func (s *WorkflowLoginFlowStep) GetOneOf() ([]WorkflowObject, bool) {
	switch s.Type {
	case WorkflowLoginFlowStepTypeIdentify:
		fallthrough
	case WorkflowLoginFlowStepTypeAuthenticate:
		out := make([]WorkflowObject, len(s.OneOf))
		for i, v := range s.OneOf {
			v := v
			out[i] = v
		}
		return out, true
	default:
		return nil, false
	}
}

type WorkflowLoginFlowOneOf struct {
	// Identification is specific to identify.
	Identification WorkflowIdentificationMethod `json:"identification,omitempty"`

	// Authentication is specific to authenticate.
	Authentication WorkflowAuthenticationMethod `json:"authentication,omitempty"`
	// TargetStep is specific to authenticate.
	TargetStep string `json:"target_step,omitempty"`

	// Steps are common.
	Steps []*WorkflowLoginFlowStep `json:"steps,omitempty"`
}

func (f *WorkflowLoginFlowOneOf) GetID() (string, bool) {
	return "", false
}

func (f *WorkflowLoginFlowOneOf) GetSteps() ([]WorkflowObject, bool) {
	out := make([]WorkflowObject, len(f.Steps))
	for i, v := range f.Steps {
		v := v
		out[i] = v
	}
	return out, true
}

func (f *WorkflowLoginFlowOneOf) GetOneOf() ([]WorkflowObject, bool) {
	return nil, false
}

type WorkflowSignupLoginFlow struct {
	ID    string                         `json:"id,omitempty"`
	Steps []*WorkflowSignupLoginFlowStep `json:"steps,omitempty"`
}

func (f *WorkflowSignupLoginFlow) GetID() (string, bool) {
	return f.ID, true
}

func (f *WorkflowSignupLoginFlow) GetSteps() ([]WorkflowObject, bool) {
	out := make([]WorkflowObject, len(f.Steps))
	for i, v := range f.Steps {
		v := v
		out[i] = v
	}
	return out, true
}

func (f *WorkflowSignupLoginFlow) GetOneOf() ([]WorkflowObject, bool) {
	return nil, false
}

type WorkflowSignupLoginFlowStep struct {
	ID    string                          `json:"id,omitempty"`
	Type  WorkflowSignupLoginFlowStepType `json:"type,omitempty"`
	OneOf []*WorkflowSignupLoginFlowOneOf `json:"one_of,omitempty"`
}

func (s *WorkflowSignupLoginFlowStep) GetID() (string, bool) {
	return s.ID, true
}

func (s *WorkflowSignupLoginFlowStep) GetSteps() ([]WorkflowObject, bool) {
	return nil, false
}

func (s *WorkflowSignupLoginFlowStep) GetOneOf() ([]WorkflowObject, bool) {
	switch s.Type {
	case WorkflowSignupLoginFlowStepTypeIdentify:
		out := make([]WorkflowObject, len(s.OneOf))
		for i, v := range s.OneOf {
			v := v
			out[i] = v
		}
		return out, true
	default:
		return nil, false
	}
}

type WorkflowSignupLoginFlowStepType string

const (
	WorkflowSignupLoginFlowStepTypeIdentify WorkflowSignupLoginFlowStepType = "identify"
)

type WorkflowSignupLoginFlowOneOf struct {
	Identification WorkflowIdentificationMethod `json:"identification,omitempty"`
	SignupFlow     string                       `json:"signup_flow,omitempty"`
	LoginFlow      string                       `json:"login_flow,omitempty"`
}

func (s *WorkflowSignupLoginFlowOneOf) GetID() (string, bool) {
	return "", false
}

func (s *WorkflowSignupLoginFlowOneOf) GetSteps() ([]WorkflowObject, bool) {
	return nil, false
}

func (s *WorkflowSignupLoginFlowOneOf) GetOneOf() ([]WorkflowObject, bool) {
	return nil, false
}

type WorkflowReauthFlow struct {
	ID    string                    `json:"id,omitempty"`
	Steps []*WorkflowReauthFlowStep `json:"steps,omitempty"`
}

func (f *WorkflowReauthFlow) GetID() (string, bool) {
	return f.ID, true
}

func (f *WorkflowReauthFlow) GetSteps() ([]WorkflowObject, bool) {
	out := make([]WorkflowObject, len(f.Steps))
	for i, v := range f.Steps {
		v := v
		out[i] = v
	}
	return out, true
}

func (f *WorkflowReauthFlow) GetOneOf() ([]WorkflowObject, bool) {
	return nil, false
}

type WorkflowReauthFlowStepType string

const (
	WorkflowReauthFlowStepTypeAuthenticate WorkflowReauthFlowStepType = "authenticate"
)

type WorkflowReauthFlowStep struct {
	ID   string                     `json:"id,omitempty"`
	Type WorkflowReauthFlowStepType `json:"type,omitempty"`

	// Optional is relevant when Type is authenticate.
	Optional *bool `json:"optional,omitempty"`

	// OneOf is relevant when Type is authenticate.
	OneOf []*WorkflowReauthFlowOneOf `json:"one_of,omitempty"`
}

func (s *WorkflowReauthFlowStep) GetID() (string, bool) {
	return s.ID, true
}

func (s *WorkflowReauthFlowStep) GetSteps() ([]WorkflowObject, bool) {
	return nil, false
}

func (s *WorkflowReauthFlowStep) GetOneOf() ([]WorkflowObject, bool) {
	switch s.Type {
	case WorkflowReauthFlowStepTypeAuthenticate:
		out := make([]WorkflowObject, len(s.OneOf))
		for i, v := range s.OneOf {
			v := v
			out[i] = v
		}
		return out, true
	default:
		return nil, false
	}
}

type WorkflowReauthFlowOneOf struct {
	Authentication WorkflowAuthenticationMethod `json:"authentication,omitempty"`
	TargetStep     string                       `json:"target_step,omitempty"`
	Steps          []*WorkflowReauthFlowStep    `json:"steps,omitempty"`
}

func (f *WorkflowReauthFlowOneOf) GetID() (string, bool) {
	return "", false
}

func (f *WorkflowReauthFlowOneOf) GetSteps() ([]WorkflowObject, bool) {
	out := make([]WorkflowObject, len(f.Steps))
	for i, v := range f.Steps {
		v := v
		out[i] = v
	}
	return out, true
}

func (f *WorkflowReauthFlowOneOf) GetOneOf() ([]WorkflowObject, bool) {
	return nil, false
}
