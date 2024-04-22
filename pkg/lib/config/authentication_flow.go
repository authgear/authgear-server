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
		"promote_flows": {
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
		},
		"account_recovery_flows": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/AuthenticationFlowAccountRecoveryFlow" }
		},
		"rate_limits": { "$ref": "#/$defs/AuthenticationFlowRateLimitsConfig" },
		"default_account_linking": { "$ref": "#/$defs/AuthenticationFlowAccountLinking" }
	}
}
`)

var _ = Schema.Add("AuthenticationFlowObjectName", `
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
	"required": ["name", "steps"],
	"properties": {
		"name": { "$ref": "#/$defs/AuthenticationFlowObjectName" },
		"account_linking": { "$ref": "#/$defs/AuthenticationFlowAccountLinking" },
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
		"name": { "$ref": "#/$defs/AuthenticationFlowObjectName" },
		"type": {
			"type": "string",
			"enum": [
				"identify",
				"create_authenticator",
				"verify",
				"fill_in_user_profile",
				"view_recovery_code",
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
					"type": { "const": "create_authenticator" }
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
					"target_step": { "$ref": "#/$defs/AuthenticationFlowObjectName" }
				}
			}
		},
		{
			"if": {
				"required": ["type"],
				"properties": {
					"type": { "const": "fill_in_user_profile" }
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
		"target_step": { "$ref": "#/$defs/AuthenticationFlowObjectName" },
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
	"required": ["name", "steps"],
	"properties": {
		"name": { "$ref": "#/$defs/AuthenticationFlowObjectName" },
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
		"name": { "$ref": "#/$defs/AuthenticationFlowObjectName" },
		"type": {
			"type": "string",
			"enum": [
				"identify",
				"authenticate",
				"check_account_status",
				"terminate_other_sessions",
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
					"target_step": { "$ref": "#/$defs/AuthenticationFlowObjectName" }
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
		"target_step": { "$ref": "#/$defs/AuthenticationFlowObjectName" },
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
	"required": ["name", "steps"],
	"properties": {
		"name": { "$ref": "#/$defs/AuthenticationFlowObjectName" },
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
		"name": { "$ref": "#/$defs/AuthenticationFlowObjectName" },
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
		"signup_flow": { "$ref": "#/$defs/AuthenticationFlowObjectName" },
		"login_flow": { "$ref": "#/$defs/AuthenticationFlowObjectName" }
	}
}
`)

var _ = Schema.Add("AuthenticationFlowReauthFlow", `
{
	"type": "object",
	"required": ["name", "steps"],
	"properties": {
		"name": { "$ref": "#/$defs/AuthenticationFlowObjectName" },
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
		"name": { "$ref": "#/$defs/AuthenticationFlowObjectName" },
		"type": {
			"type": "string",
			"enum": [
				"identify",
				"authenticate"
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
						"items": { "$ref": "#/$defs/AuthenticationFlowReauthFlowIdentify" }
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
						"items": { "$ref": "#/$defs/AuthenticationFlowReauthFlowAuthenticate" }
					}
				}
			}
		}
	]
}
`)

var _ = Schema.Add("AuthenticationFlowReauthFlowIdentify", `
{
	"type": "object",
	"required": ["identification"],
	"properties": {
		"identification": {
			"type": "string",
			"enum": [
				"id_token"
			]
		},
		"steps": {
			"type": "array",
			"items": { "$ref": "#/$defs/AuthenticationFlowReauthFlowStep" }
		}
	}
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
		"steps": {
			"type": "array",
			"items": { "$ref": "#/$defs/AuthenticationFlowReauthFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("AuthenticationFlowAccountRecoveryFlow", `
{
	"type": "object",
	"required": ["name", "steps"],
	"properties": {
		"name": { "$ref": "#/$defs/AuthenticationFlowObjectName" },
		"steps": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/AuthenticationFlowAccountRecoveryFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("AuthenticationFlowAccountRecoveryFlowStep", `
{
	"type": "object",
	"required": ["type"],
	"properties": {
		"name": { "$ref": "#/$defs/AuthenticationFlowObjectName" },
		"type": {
			"type": "string",
			"enum": [
				"identify",
				"select_destination",
				"verify_account_recovery_code",
				"reset_password"
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
						"items": { "$ref": "#/$defs/AuthenticationFlowAccountRecoveryFlowOneOf" }
					}
				}
			}
		},
		{
			"if": {
				"required": ["type"],
				"properties": {
					"type": { "const": "select_destination" }
				}
			},
			"then": {
				"properties": {
					"enumerate_destinations": {
						"type": "boolean"
					},
					"allowed_channels": {
						"type": "array",
						"items": { "$ref": "#/$defs/AccountRecoveryChannel" }
					}
				}
			}
		}
	]
}
`)

var _ = Schema.Add("AuthenticationFlowAccountRecoveryFlowOneOf", `
{
	"type": "object",
	"required": ["identification"],
	"properties": {
		"identification": { "$ref": "#/$defs/AuthenticationFlowAccountRecoveryIdentification" },
		"on_failure": { "type": "string", "enum": [ "error", "ignore"] },
		"steps": {
			"type": "array",
			"minItems": 1,
			"items": { "$ref": "#/$defs/AuthenticationFlowAccountRecoveryFlowStep" }
		}
	}
}
`)

var _ = Schema.Add("AuthenticationFlowAccountRecoveryIdentification", `
{
	"type": "string",
	"enum": [
		"email",
		"phone"
	]
}
`)

var _ = Schema.Add("AuthenticationFlowAccountLinking", `
{
	"type": "object",
	"required": [],
	"properties": {
		"oauth": {
			"type": "array",
			"items": { "$ref": "#/$defs/AccountLinkingOAuth" }
		}
	}
}
`)

var _ = Schema.Add("AccountLinkingOAuth", `
{
	"type": "object",
	"required": ["alias", "oauth_claim", "user_profile", "action"],
	"properties": {
		"alias": { "type": "string" },
		"oauth_claim": { "$ref": "#/$defs/AccountLinkingJSONPointer" },
		"user_profile": { "$ref": "#/$defs/AccountLinkingJSONPointer" },
		"action": { "$ref": "#/$defs/AccountLinkingOAuthAction" }
	},
	"allOf": [
		{
			"if": {
				"required": ["action"],
				"properties": {
					"action": {
						"type": "string",
						"enum": ["login_and_link"]
					}
				}
			},
			"then": {
				"required": ["login_flow"],
				"properties": {
					"login_flow": {
						"type": "string"
					}
				}
			}
		}
	]
}
`)

var _ = Schema.Add("AccountLinkingOAuthAction", `
{
	"type": "string",
	"enum": [
		"error",
		"login_and_link"
	]
}
`)

var _ = Schema.Add("AccountLinkingJSONPointer", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"pointer": {
			"type": "string",
			"format": "json-pointer"
		}
	}
}
`)

type AuthenticationFlowObject interface {
	IsFlowObject()
}

type AuthenticationFlowObjectFlowRoot interface {
	AuthenticationFlowObject
	GetName() string
	GetSteps() []AuthenticationFlowObject
}

type AuthenticationFlowType string

const (
	AuthenticationFlowTypeSignup          AuthenticationFlowType = "signup"
	AuthenticationFlowTypePromote         AuthenticationFlowType = "promote"
	AuthenticationFlowTypeLogin           AuthenticationFlowType = "login"
	AuthenticationFlowTypeSignupLogin     AuthenticationFlowType = "signup_login"
	AuthenticationFlowTypeReauth          AuthenticationFlowType = "reauth"
	AuthenticationFlowTypeAccountRecovery AuthenticationFlowType = "account_recovery"
)

type AuthenticationFlowObjectFlowStep interface {
	AuthenticationFlowObject
	GetName() string
	GetType() AuthenticationFlowStepType
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
	AuthenticationFlowIdentificationIDToken  AuthenticationFlowIdentification = "id_token"
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

type AuthenticationFlowStepType string

const (
	AuthenticationFlowStepTypeIdentify                  AuthenticationFlowStepType = "identify"
	AuthenticationFlowStepTypeAuthenticate              AuthenticationFlowStepType = "authenticate"
	AuthenticationFlowStepTypeCreateAuthenticator       AuthenticationFlowStepType = "create_authenticator"
	AuthenticationFlowStepTypeVerify                    AuthenticationFlowStepType = "verify"
	AuthenticationFlowStepTypeFillInUserProfile         AuthenticationFlowStepType = "fill_in_user_profile"
	AuthenticationFlowStepTypeViewRecoveryCode          AuthenticationFlowStepType = "view_recovery_code"
	AuthenticationFlowStepTypePromptCreatePasskey       AuthenticationFlowStepType = "prompt_create_passkey"
	AuthenticationFlowStepTypeTerminateOtherSessions    AuthenticationFlowStepType = "terminate_other_sessions"
	AuthenticationFlowStepTypeCheckAccountStatus        AuthenticationFlowStepType = "check_account_status"
	AuthenticationFlowStepTypeChangePassword            AuthenticationFlowStepType = "change_password"
	AuthenticationFlowStepTypeSelectDestination         AuthenticationFlowStepType = "select_destination"
	AuthenticationFlowStepTypeVerifyAccountRecoveryCode AuthenticationFlowStepType = "verify_account_recovery_code"
	AuthenticationFlowStepTypeResetPassword             AuthenticationFlowStepType = "reset_password"
)

type AuthenticationFlowConfig struct {
	SignupFlows []*AuthenticationFlowSignupFlow `json:"signup_flows,omitempty"`
	// PromoteFlows is intentionally of type AuthenticationFlowSignupFlow
	PromoteFlows         []*AuthenticationFlowSignupFlow          `json:"promote_flows,omitempty"`
	LoginFlows           []*AuthenticationFlowLoginFlow           `json:"login_flows,omitempty"`
	SignupLoginFlows     []*AuthenticationFlowSignupLoginFlow     `json:"signup_login_flows,omitempty"`
	ReauthFlows          []*AuthenticationFlowReauthFlow          `json:"reauth_flows,omitempty"`
	AccountRecoveryFlows []*AuthenticationFlowAccountRecoveryFlow `json:"account_recovery_flows,omitempty"`

	RateLimits            *AuthenticationFlowRateLimitsConfig `json:"rate_limits,omitempty"`
	DefaultAccountLinking *AuthenticationFlowAccountLinking   `json:"default_account_linking,omitempty"`
}

type AuthenticationFlowSignupFlow struct {
	Name           string                              `json:"name,omitempty"`
	AccountLinking *AuthenticationFlowAccountLinking   `json:"account_linking,omitempty"`
	Steps          []*AuthenticationFlowSignupFlowStep `json:"steps,omitempty"`
}

var _ AuthenticationFlowObjectFlowRoot = &AuthenticationFlowSignupFlow{}

func (f *AuthenticationFlowSignupFlow) IsFlowObject()   {}
func (f *AuthenticationFlowSignupFlow) GetName() string { return f.Name }
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
	AuthenticationFlowSignupFlowStepTypeIdentify            = AuthenticationFlowSignupFlowStepType(AuthenticationFlowStepTypeIdentify)
	AuthenticationFlowSignupFlowStepTypeCreateAuthenticator = AuthenticationFlowSignupFlowStepType(AuthenticationFlowStepTypeCreateAuthenticator)
	AuthenticationFlowSignupFlowStepTypeVerify              = AuthenticationFlowSignupFlowStepType(AuthenticationFlowStepTypeVerify)
	AuthenticationFlowSignupFlowStepTypeFillInUserProfile   = AuthenticationFlowSignupFlowStepType(AuthenticationFlowStepTypeFillInUserProfile)
	AuthenticationFlowSignupFlowStepTypeViewRecoveryCode    = AuthenticationFlowSignupFlowStepType(AuthenticationFlowStepTypeViewRecoveryCode)
	AuthenticationFlowSignupFlowStepTypePromptCreatePasskey = AuthenticationFlowSignupFlowStepType(AuthenticationFlowStepTypePromptCreatePasskey)
)

type AuthenticationFlowSignupFlowStep struct {
	Name string                               `json:"name,omitempty"`
	Type AuthenticationFlowSignupFlowStepType `json:"type,omitempty"`

	// OneOf is relevant when Type is identify or authenticate.
	OneOf []*AuthenticationFlowSignupFlowOneOf `json:"one_of,omitempty"`
	// TargetStep is relevant when Type is verify.
	TargetStep string `json:"target_step,omitempty"`
	// UserProfile is relevant when Type is fill_in_user_profile.
	UserProfile []*AuthenticationFlowSignupFlowUserProfile `json:"user_profile,omitempty"`
}

var _ AuthenticationFlowObjectFlowStep = &AuthenticationFlowSignupFlowStep{}

func (s *AuthenticationFlowSignupFlowStep) IsFlowObject()   {}
func (s *AuthenticationFlowSignupFlowStep) GetName() string { return s.Name }
func (s *AuthenticationFlowSignupFlowStep) GetType() AuthenticationFlowStepType {
	return AuthenticationFlowStepType(s.Type)
}
func (s *AuthenticationFlowSignupFlowStep) GetOneOf() []AuthenticationFlowObject {
	switch s.Type {
	case AuthenticationFlowSignupFlowStepTypeIdentify:
		fallthrough
	case AuthenticationFlowSignupFlowStepTypeCreateAuthenticator:
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
	// VerificationRequired is specific to OOB.
	VerificationRequired *bool `json:"verification_required,omitempty"`

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

func (f *AuthenticationFlowSignupFlowOneOf) IsVerificationRequired() bool {
	// If it is unspecified (i.e. nil), then verification is required.
	return f.VerificationRequired == nil || *f.VerificationRequired
}

type AuthenticationFlowSignupFlowUserProfile struct {
	Pointer  string `json:"pointer,omitempty"`
	Required bool   `json:"required,omitempty"`
}

type AuthenticationFlowLoginFlow struct {
	Name  string                             `json:"name,omitempty"`
	Steps []*AuthenticationFlowLoginFlowStep `json:"steps,omitempty"`
}

var _ AuthenticationFlowObjectFlowRoot = &AuthenticationFlowLoginFlow{}

func (f *AuthenticationFlowLoginFlow) IsFlowObject() {}

func (f *AuthenticationFlowLoginFlow) GetName() string {
	return f.Name
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
	AuthenticationFlowLoginFlowStepTypeIdentify               = AuthenticationFlowLoginFlowStepType(AuthenticationFlowStepTypeIdentify)
	AuthenticationFlowLoginFlowStepTypeAuthenticate           = AuthenticationFlowLoginFlowStepType(AuthenticationFlowStepTypeAuthenticate)
	AuthenticationFlowLoginFlowStepTypeCheckAccountStatus     = AuthenticationFlowLoginFlowStepType(AuthenticationFlowStepTypeCheckAccountStatus)
	AuthenticationFlowLoginFlowStepTypeTerminateOtherSessions = AuthenticationFlowLoginFlowStepType(AuthenticationFlowStepTypeTerminateOtherSessions)
	AuthenticationFlowLoginFlowStepTypeChangePassword         = AuthenticationFlowLoginFlowStepType(AuthenticationFlowStepTypeChangePassword)
	AuthenticationFlowLoginFlowStepTypePromptCreatePasskey    = AuthenticationFlowLoginFlowStepType(AuthenticationFlowStepTypePromptCreatePasskey)
)

type AuthenticationFlowLoginFlowStep struct {
	Name string                              `json:"name,omitempty"`
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
func (s *AuthenticationFlowLoginFlowStep) GetName() string { return s.Name }
func (s *AuthenticationFlowLoginFlowStep) GetType() AuthenticationFlowStepType {
	return AuthenticationFlowStepType(s.Type)
}

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

func (s *AuthenticationFlowLoginFlowStep) IsOptional() bool {
	if s.Optional != nil && *s.Optional {
		return true
	}
	return false
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
	Name  string                                   `json:"name,omitempty"`
	Steps []*AuthenticationFlowSignupLoginFlowStep `json:"steps,omitempty"`
}

var _ AuthenticationFlowObjectFlowRoot = &AuthenticationFlowSignupLoginFlow{}

func (f *AuthenticationFlowSignupLoginFlow) IsFlowObject()   {}
func (f *AuthenticationFlowSignupLoginFlow) GetName() string { return f.Name }

func (f *AuthenticationFlowSignupLoginFlow) GetSteps() []AuthenticationFlowObject {
	out := make([]AuthenticationFlowObject, len(f.Steps))
	for i, v := range f.Steps {
		v := v
		out[i] = v
	}
	return out
}

type AuthenticationFlowSignupLoginFlowStep struct {
	Name  string                                    `json:"name,omitempty"`
	Type  AuthenticationFlowSignupLoginFlowStepType `json:"type,omitempty"`
	OneOf []*AuthenticationFlowSignupLoginFlowOneOf `json:"one_of,omitempty"`
}

var _ AuthenticationFlowObjectFlowStep = &AuthenticationFlowSignupLoginFlowStep{}

func (s *AuthenticationFlowSignupLoginFlowStep) IsFlowObject()   {}
func (s *AuthenticationFlowSignupLoginFlowStep) GetName() string { return s.Name }
func (s *AuthenticationFlowSignupLoginFlowStep) GetType() AuthenticationFlowStepType {
	return AuthenticationFlowStepType(s.Type)
}

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
	AuthenticationFlowSignupLoginFlowStepTypeIdentify = AuthenticationFlowSignupLoginFlowStepType(AuthenticationFlowStepTypeIdentify)
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
	Name  string                              `json:"name,omitempty"`
	Steps []*AuthenticationFlowReauthFlowStep `json:"steps,omitempty"`
}

var _ AuthenticationFlowObjectFlowRoot = &AuthenticationFlowReauthFlow{}

func (f *AuthenticationFlowReauthFlow) IsFlowObject()   {}
func (f *AuthenticationFlowReauthFlow) GetName() string { return f.Name }

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
	AuthenticationFlowReauthFlowStepTypeIdentify     = AuthenticationFlowReauthFlowStepType(AuthenticationFlowStepTypeIdentify)
	AuthenticationFlowReauthFlowStepTypeAuthenticate = AuthenticationFlowReauthFlowStepType(AuthenticationFlowStepTypeAuthenticate)
)

type AuthenticationFlowReauthFlowStep struct {
	Name string                               `json:"name,omitempty"`
	Type AuthenticationFlowReauthFlowStepType `json:"type,omitempty"`

	// OneOf is relevant when Type is identify or authenticate.
	OneOf []*AuthenticationFlowReauthFlowOneOf `json:"one_of,omitempty"`
}

var _ AuthenticationFlowObjectFlowStep = &AuthenticationFlowReauthFlowStep{}

func (s *AuthenticationFlowReauthFlowStep) IsFlowObject()   {}
func (s *AuthenticationFlowReauthFlowStep) GetName() string { return s.Name }
func (s *AuthenticationFlowReauthFlowStep) GetType() AuthenticationFlowStepType {
	return AuthenticationFlowStepType(s.Type)
}

func (s *AuthenticationFlowReauthFlowStep) GetOneOf() []AuthenticationFlowObject {
	switch s.Type {
	case AuthenticationFlowReauthFlowStepTypeIdentify:
		fallthrough
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
	// Identification is specific to identify.
	Identification AuthenticationFlowIdentification `json:"identification,omitempty"`

	// Authentication is specific to authenticate.
	Authentication AuthenticationFlowAuthentication `json:"authentication,omitempty"`

	// Steps are common.
	Steps []*AuthenticationFlowReauthFlowStep `json:"steps,omitempty"`
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
		Identification: f.Identification,
		Authentication: f.Authentication,
	}
}

type AuthenticationFlowAccountRecoveryFlow struct {
	Name  string                                       `json:"name,omitempty"`
	Steps []*AuthenticationFlowAccountRecoveryFlowStep `json:"steps,omitempty"`
}

var _ AuthenticationFlowObjectFlowRoot = &AuthenticationFlowAccountRecoveryFlow{}

func (f *AuthenticationFlowAccountRecoveryFlow) IsFlowObject() {}

func (f *AuthenticationFlowAccountRecoveryFlow) GetName() string {
	return f.Name
}

func (f *AuthenticationFlowAccountRecoveryFlow) GetSteps() []AuthenticationFlowObject {
	out := make([]AuthenticationFlowObject, len(f.Steps))
	for i, v := range f.Steps {
		v := v
		out[i] = v
	}
	return out
}

type AccountRecoveryCodeChannel string

const (
	AccountRecoveryCodeChannelSMS      AccountRecoveryCodeChannel = "sms"
	AccountRecoveryCodeChannelEmail    AccountRecoveryCodeChannel = "email"
	AccountRecoveryCodeChannelWhatsapp AccountRecoveryCodeChannel = "whatsapp"
)

type AccountRecoveryCodeForm string

const (
	AccountRecoveryCodeFormLink AccountRecoveryCodeForm = "link"
	AccountRecoveryCodeFormCode AccountRecoveryCodeForm = "code"
)

type AccountRecoveryChannel struct {
	Channel AccountRecoveryCodeChannel `json:"channel,omitempty"`
	OTPForm AccountRecoveryCodeForm    `json:"otp_form,omitempty"`
}

func GetAllAccountRecoveryChannel() []*AccountRecoveryChannel {
	return []*AccountRecoveryChannel{
		{
			Channel: AccountRecoveryCodeChannelEmail,
			OTPForm: AccountRecoveryCodeFormLink,
		},
		{
			Channel: AccountRecoveryCodeChannelEmail,
			OTPForm: AccountRecoveryCodeFormCode,
		},
		{
			Channel: AccountRecoveryCodeChannelSMS,
			OTPForm: AccountRecoveryCodeFormCode,
		},
		{
			Channel: AccountRecoveryCodeChannelSMS,
			OTPForm: AccountRecoveryCodeFormLink,
		},
		{
			Channel: AccountRecoveryCodeChannelWhatsapp,
			OTPForm: AccountRecoveryCodeFormCode,
		},
	}
}

type AuthenticationFlowAccountRecoveryFlowStep struct {
	Name string                                    `json:"name,omitempty"`
	Type AuthenticationFlowAccountRecoveryFlowType `json:"type,omitempty"`
	// OneOf is relevant when Type is identify.
	OneOf []*AuthenticationFlowAccountRecoveryFlowOneOf `json:"one_of,omitempty"`
	// EnumerateDestinations is specific to select_destination.
	EnumerateDestinations bool `json:"enumerate_destinations,omitempty"`
	// AllowedChannels is specific to select_destination.
	AllowedChannels []*AccountRecoveryChannel `json:"allowed_channels,omitempty"`
}

var _ AuthenticationFlowObjectFlowStep = &AuthenticationFlowAccountRecoveryFlowStep{}

func (s *AuthenticationFlowAccountRecoveryFlowStep) IsFlowObject()   {}
func (s *AuthenticationFlowAccountRecoveryFlowStep) GetName() string { return s.Name }
func (s *AuthenticationFlowAccountRecoveryFlowStep) GetType() AuthenticationFlowStepType {
	return AuthenticationFlowStepType(s.Type)
}
func (s *AuthenticationFlowAccountRecoveryFlowStep) GetOneOf() []AuthenticationFlowObject {
	switch s.Type {
	case AuthenticationFlowAccountRecoveryFlowTypeIdentify:
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

type AuthenticationFlowAccountRecoveryFlowType string

const (
	AuthenticationFlowAccountRecoveryFlowTypeIdentify                  = AuthenticationFlowAccountRecoveryFlowType(AuthenticationFlowStepTypeIdentify)
	AuthenticationFlowAccountRecoveryFlowTypeSelectDestination         = AuthenticationFlowAccountRecoveryFlowType(AuthenticationFlowStepTypeSelectDestination)
	AuthenticationFlowAccountRecoveryFlowTypeVerifyAccountRecoveryCode = AuthenticationFlowAccountRecoveryFlowType(AuthenticationFlowStepTypeVerifyAccountRecoveryCode)
	AuthenticationFlowAccountRecoveryFlowTypeResetPassword             = AuthenticationFlowAccountRecoveryFlowType(AuthenticationFlowStepTypeResetPassword)
)

type AuthenticationFlowAccountRecoveryFlowOneOf struct {
	Identification AuthenticationFlowAccountRecoveryIdentification          `json:"identification,omitempty"`
	OnFailure      AuthenticationFlowAccountRecoveryIdentificationOnFailure `json:"on_failure,omitempty"`
	Steps          []*AuthenticationFlowAccountRecoveryFlowStep             `json:"steps,omitempty"`
}

var _ AuthenticationFlowObjectFlowBranch = &AuthenticationFlowAccountRecoveryFlowOneOf{}

func (f *AuthenticationFlowAccountRecoveryFlowOneOf) IsFlowObject() {}
func (f *AuthenticationFlowAccountRecoveryFlowOneOf) GetSteps() []AuthenticationFlowObject {
	out := make([]AuthenticationFlowObject, len(f.Steps))
	for i, v := range f.Steps {
		v := v
		out[i] = v
	}
	return out
}
func (f *AuthenticationFlowAccountRecoveryFlowOneOf) GetBranchInfo() AuthenticationFlowObjectFlowBranchInfo {
	return AuthenticationFlowObjectFlowBranchInfo{
		Identification: AuthenticationFlowIdentification(f.Identification),
	}
}

type AuthenticationFlowAccountRecoveryIdentification AuthenticationFlowIdentification

const (
	AuthenticationFlowAccountRecoveryIdentificationEmail = AuthenticationFlowAccountRecoveryIdentification(AuthenticationFlowIdentificationEmail)
	AuthenticationFlowAccountRecoveryIdentificationPhone = AuthenticationFlowAccountRecoveryIdentification(AuthenticationFlowIdentificationPhone)
)

func (i AuthenticationFlowAccountRecoveryIdentification) AuthenticationFlowIdentification() AuthenticationFlowIdentification {
	return AuthenticationFlowIdentification(i)
}

type AuthenticationFlowAccountRecoveryIdentificationOnFailure string

const (
	AuthenticationFlowAccountRecoveryIdentificationOnFailureError  = AuthenticationFlowAccountRecoveryIdentificationOnFailure("error")
	AuthenticationFlowAccountRecoveryIdentificationOnFailureIgnore = AuthenticationFlowAccountRecoveryIdentificationOnFailure("ignore")
)

type AuthenticationFlowAccountLinking struct {
	OAuth []*AccountLinkingOAuth `json:"oauth,omitempty"`
}

type AccountLinkingOAuth struct {
	Alias       string                    `json:"alias,omitempty"`
	OAuthClaim  AccountLinkingJSONPointer `json:"oauth_claim,omitempty"`
	UserProfile AccountLinkingJSONPointer `json:"user_profile,omitempty"`
	Action      AccountLinkingOAuthAction `json:"action,omitempty"`

	// login_flow is only relevant if action is "login_and_link"
	LoginFlow string `json:"login_flow,omitempty"`
}

type AccountLinkingOAuthAction string

const (
	AccountLinkingOAuthActionError        AccountLinkingOAuthAction = "error"
	AccountLinkingOAuthActionLoginAndLink AccountLinkingOAuthAction = "login_andLink"
)

type AccountLinkingJSONPointer struct {
	Pointer string `json:"pointer,omitempty"`
}

func init() {
	accountRecoveryChannelsOneOf := ""
	addAccountRecoveryChannel := func(channel AccountRecoveryCodeChannel, otpForm AccountRecoveryCodeForm) {
		newChannel := fmt.Sprintf(`
		{
			"properties": {
				"channel": { "const": "%s" },
				"otp_form": { "const": "%s" }
			},
			"required": ["channel", "otp_form"]
		}`, channel, otpForm)
		if accountRecoveryChannelsOneOf == "" {
			accountRecoveryChannelsOneOf = newChannel
		} else {
			accountRecoveryChannelsOneOf = fmt.Sprintf(`
		%s,
		%s`, accountRecoveryChannelsOneOf, newChannel)
		}
	}
	for _, c := range GetAllAccountRecoveryChannel() {
		addAccountRecoveryChannel(c.Channel, c.OTPForm)
	}
	schema := fmt.Sprintf(`
	{
		"type": "object",
		"oneOf": [
%s
		]
	}`, accountRecoveryChannelsOneOf)
	var _ = Schema.Add("AccountRecoveryChannel", schema)
}
