package event

import (
	"context"
	"fmt"
	"io"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/slice"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var HookResponseSchema = validation.NewSimpleSchema(`
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
							"items": {
								"type": "string"
							}
						}
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

type HookResponse struct {
	IsAllowed   bool         `json:"is_allowed"`
	Title       string       `json:"title,omitempty"`
	Reason      string       `json:"reason,omitempty"`
	Mutations   Mutations    `json:"mutations,omitempty"`
	Constraints *Constraints `json:"constraints,omitempty"`
}

var supportedAMRConstraints = []string{model.AMRMFA, model.AMROTP, model.AMRPWD, model.AMRSMS}

type Constraints struct {
	AMR []string `json:"amr,omitempty"`
}

func (c *Constraints) Validate() error {
	for _, amr := range c.AMR {
		if !slice.ContainsString(supportedAMRConstraints, amr) {
			return fmt.Errorf("unsupported amr constraint %s", amr)
		}
	}
	return nil
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
	if resp.Constraints != nil {
		err := resp.Constraints.Validate()
		if err != nil {
			return nil, err
		}
	}
	return &resp, nil
}
