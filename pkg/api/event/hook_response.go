package event

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var HookResponseSchema *validation.MultipartSchema

func init() {
	var supportedAMRConstraints = []string{model.AMRMFA, model.AMROTP, model.AMRPWD, model.AMRSMS}
	supportedAMRConstraintsJSON, err := json.Marshal(supportedAMRConstraints)
	if err != nil {
		panic(err)
	}
	HookResponseSchema = validation.NewMultipartSchema("HookResponseSchema")
	_ = HookResponseSchema.Add("AMRConstraint", fmt.Sprintf(`
{
	"type": "string",
	"enum": %s
}
`, string(supportedAMRConstraintsJSON)))

	_ = HookResponseSchema.Add("BotProtectionRiskMode", `
{
	"type": "string",
	"enum": ["never", "always"]
}
`)

	_ = HookResponseSchema.Add("HookResponseSchema", `
{
	"oneOf": [
		{
			"type": "object",
			"additionalProperties": false,
			"properties": {
				"is_allowed": { "const": true },
				"mutations": {
					"type": "object",
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
						}
					}
				},
				"constraints": {
					"type": "object",
					"properties": {
						"amr": {
							"type": "array",
							"items": { "$ref": "#/$defs/AMRConstraint" }
						}
					}
				},
				"bot_protection": {
					"type": "object",
					"properties": {
						"mode": { "$ref": "#/$defs/BotProtectionRiskMode" }
					}
				}
			},
			"required": ["is_allowed"]
		},
		{
			"type": "object",
			"additionalProperties": false,
			"properties": {
				"is_allowed": { "const": false },
				"title": { "type": "string" },
				"reason": { "type": "string" }
			},
			"required": ["is_allowed"]
		}
	]
}
`)

	HookResponseSchema.Instantiate()
}

type HookResponse struct {
	IsAllowed     bool                       `json:"is_allowed"`
	Title         string                     `json:"title,omitempty"`
	Reason        string                     `json:"reason,omitempty"`
	Mutations     Mutations                  `json:"mutations,omitempty"`
	Constraints   *Constraints               `json:"constraints,omitempty"`
	BotProtection *BotProtectionRequirements `json:"bot_protection,omitempty"`
}

type Constraints struct {
	AMR []string `json:"amr,omitempty"`
}

type BotProtectionRequirements struct {
	Mode config.BotProtectionRiskMode `json:"mode,omitempty"`
}

type Mutations struct {
	User UserMutations `json:"user,omitempty"`
	JWT  JWTMutations  `json:"jwt,omitempty"`
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

func ParseHookResponse(ctx context.Context, r io.Reader) (*HookResponse, error) {
	var resp HookResponse
	if err := HookResponseSchema.Validator().Parse(ctx, r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
