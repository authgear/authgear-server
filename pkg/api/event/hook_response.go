package event

import (
	"io"

	"github.com/authgear/authgear-server/pkg/util/validation"
)

/*
	@ID HookResponse
	@Response
		Validation result of the event, and optionally mutate the user object.

		@JSONSchema
		@JSONExample Allowed - Allow operation
			{
				"is_allowed": true
			}
		@JSONExample Disallowed - Disallow operation with reason
			{
				"is_allowed": false,
				"title": "Validation failure",
				"reason": "Username is not allowed"
			}
*/

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
	Title     string    `json:"title"`
	Reason    string    `json:"reason"`
	Mutations Mutations `json:"mutations"`
}

type Mutations struct {
	User UserMutations `json:"user"`
}

type UserMutations struct {
	StandardAttributes map[string]interface{} `json:"standard_attributes,omitempty"`
	CustomAttributes   map[string]interface{} `json:"custom_attributes,omitempty"`
}

func ParseHookResponse(r io.Reader) (*HookResponse, error) {
	var resp HookResponse
	if err := HookResponseSchema.Validator().Parse(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
