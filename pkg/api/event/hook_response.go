package event

import (
	"io"

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
	IsAllowed bool      `json:"is_allowed"`
	Title     string    `json:"title,omitempty"`
	Reason    string    `json:"reason,omitempty"`
	Mutations Mutations `json:"mutations,omitempty"`
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

func ParseHookResponse(r io.Reader) (*HookResponse, error) {
	var resp HookResponse
	if err := HookResponseSchema.Validator().Parse(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
