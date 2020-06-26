package event

import (
	"io"

	"github.com/skygeario/skygear-server/pkg/validation"
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
				"reason": "Validation failure",
				"data": { "fields": ["user_name"] }
			}
*/

var HookResponseSchema = validation.NewSimpleSchema(`
{
	"allOf": [
		{
			"if": {
				"properties": {
					"is_allowed": { "type": "boolean", "enum": [true] }
				}
			},
			"then": {
				"additionalProperties": false,
				"properties": {
					"is_allowed": { "type": "boolean" },
					"mutations": {
						"type": "object",
						"properties": {
							"metadata": { "type": "object" }
						}
					}
				},
				"required": ["is_allowed"]
			}
		},
		{
			"if": {
				"properties": {
					"is_allowed": { "type": "boolean", "enum": [false] }
				}
			},
			"then": {
				"additionalProperties": false,
				"properties": {
					"is_allowed": { "type": "boolean" },
					"reason": { "type": "string" },
					"data": { "type": "object" }
				},
				"required": ["is_allowed", "reason"]
			}
		}
	]
}
`)

type HookResponse struct {
	IsAllowed bool        `json:"is_allowed"`
	Reason    string      `json:"reason"`
	Data      interface{} `json:"data"`
	Mutations *Mutations  `json:"mutations"`
}

func ParseHookResponse(r io.Reader) (*HookResponse, error) {
	var resp HookResponse
	if err := HookResponseSchema.Validator().Parse(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
