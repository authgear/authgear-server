package event

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func GetBaseHookResponseSchema() *validation.MultipartSchema {
	var baseHookResponseSchema *validation.MultipartSchema = validation.NewMultipartSchema("BaseHookResponseSchema")
	var supportedAMRConstraints = []string{
		model.AMRMFA,
		model.AMROTP,
		model.AMRPWD,
		model.AMRSMS,
		model.AMRXPrimaryOOBOTPEmail,
		model.AMRXPrimaryOOBOTPSMS,
		model.AMRXPrimaryPassword,
		model.AMRXRecoveryCode,
		model.AMRXSecondaryOOBOTPEmail,
		model.AMRXSecondaryOOBOTPSMS,
		model.AMRXSecondaryPassword,
		model.AMRXSecondaryTOTP,
	}
	supportedAMRConstraintsJSON, err := json.Marshal(supportedAMRConstraints)
	if err != nil {
		panic(err)
	}
	_ = baseHookResponseSchema.Add("AMRConstraint", fmt.Sprintf(`
{
	"type": "string",
	"enum": %s
}
`, string(supportedAMRConstraintsJSON)))

	_ = baseHookResponseSchema.Add("BotProtectionRiskMode", `
{
	"type": "string",
	"enum": ["never", "always"]
}
`)

	_ = baseHookResponseSchema.Add("RateLimits", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"authentication.general": { "$ref": "#/$defs/RateLimitRequirements" },
		"authentication.account_enumeration": { "$ref": "#/$defs/RateLimitRequirements" }
	}
}
`)

	_ = baseHookResponseSchema.Add("RateLimitRequirements", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"weight": { "type": "number", "minimum": 0 }
	}
}
`)

	_ = baseHookResponseSchema.Add("BotProtectionRequirements", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"mode": { "$ref": "#/$defs/BotProtectionRiskMode" }
	}
}
`)

	_ = baseHookResponseSchema.Add("Mutations", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"user": {
			"type": "object",
			"properties": {
				"standard_attributes": {
					"type": "object"
				},
				"custom_attributes": {
					"type": "object"
				}
			}
		},
		"jwt": {
			"type": "object",
			"properties": {
				"payload": {
					"type": "object"
				}
			}
		},
		"id_token": {
			"type": "object",
			"properties": {
				"payload": {
					"type": "object"
				}
			}
		}
	}
}
`)

	_ = baseHookResponseSchema.Add("Constraints", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"amr": {
			"type": "array",
			"items": { "$ref": "#/$defs/AMRConstraint" }
		}
	}
}
`)

	_ = baseHookResponseSchema.Add("BaseHookResponseSchema", `
{
	"allOf": [
		{
			"properties": {
				"is_allowed": { "type": "boolean" }
			},
			"required": ["is_allowed"]
		},
		{
			"if": {
				"properties": {
					"is_allowed": { "const": true }
				}
			},
			"then": {
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"is_allowed": { "const": true },
					"mutations": { "$ref": "#/$defs/Mutations" },
					"constraints": { "$ref": "#/$defs/Constraints" },
					"bot_protection": { "$ref": "#/$defs/BotProtectionRequirements" },
					"rate_limits": { "$ref": "#/$defs/RateLimits" }
				}
			}
		},
		{
			"if": {
				"properties": {
					"is_allowed": { "const": false }
				}
			},
			"then": {
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"is_allowed": { "const": false },
					"title": { "type": "string" },
					"reason": { "type": "string" }
				}
			}
		}
	]
}
`)

	return baseHookResponseSchema
}

var responseSchemaValidators map[Type]*validation.SchemaValidator = map[Type]*validation.SchemaValidator{}

func RegisterResponseSchemaValidator(typ Type, v *validation.SchemaValidator) {
	responseSchemaValidators[typ] = v
}

type HookResponse struct {
	IsAllowed     bool                       `json:"is_allowed"`
	Title         string                     `json:"title,omitempty"`
	Reason        string                     `json:"reason,omitempty"`
	Mutations     Mutations                  `json:"mutations,omitempty"`
	Constraints   *Constraints               `json:"constraints,omitempty"`
	BotProtection *BotProtectionRequirements `json:"bot_protection,omitempty"`
	RateLimits    RateLimits                 `json:"rate_limits,omitempty"`
}

type Constraints struct {
	AMR []string `json:"amr,omitempty"`
}

type BotProtectionRequirements struct {
	Mode config.BotProtectionRiskMode `json:"mode,omitempty"`
}

type RateLimits map[ratelimit.RateLimit]RateLimitRequirements

type RateLimitRequirements struct {
	Weight float64 `json:"weight,omitempty"`
}

type Mutations struct {
	User    UserMutations    `json:"user,omitempty"`
	JWT     JWTMutations     `json:"jwt,omitempty"`
	IDToken IDTokenMutations `json:"id_token,omitempty"`
}

type UserMutations struct {
	StandardAttributes map[string]interface{} `json:"standard_attributes,omitempty"`
	CustomAttributes   map[string]interface{} `json:"custom_attributes,omitempty"`
	Roles              []string               `json:"roles,omitempty"`
	Groups             []string               `json:"groups,omitempty"`
}

type JWTMutations struct {
	Payload map[string]interface{} `json:"payload,omitempty"`
}

type IDTokenMutations struct {
	Payload map[string]interface{} `json:"payload,omitempty"`
}

func ParseHookResponse(ctx context.Context, eventType Type, r io.Reader) (*HookResponse, error) {
	var resp HookResponse
	if v, ok := responseSchemaValidators[eventType]; ok && v != nil {
		err := v.Parse(ctx, r, &resp)
		if err != nil {
			return nil, err
		}

		return &resp, nil
	} else {
		panic(fmt.Errorf("event %v has no response schema validators", eventType))
	}
}
