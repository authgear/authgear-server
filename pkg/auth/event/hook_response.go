package event

import (
	"io"

	"github.com/skygeario/skygear-server/pkg/core/validation"
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
const HookResponseSchema = `
{
	"$id": "#HookResponse",
	"oneOf": [
		{
			"additionalProperties": false,
			"properties": {
				"is_allowed": { "type": "boolean", "enum": [true] },
				"mutations": {
					"type": "object",
					"properties": {
						"metadata": { "type": "object" }
					}
				}
			},
			"required": ["is_allowed"]
		},
		{
			"additionalProperties": false,
			"properties": {
				"is_allowed": { "type": "boolean", "enum": [false] },
				"reason": { "type": "string" },
				"data": { "type": "object" }
			},
			"required": ["is_allowed", "reason"]
		}
	]
}
`

var (
	hookRespValidator *validation.Validator
)

func init() {
	hookRespValidator = validation.NewValidator("http://v2.skygear.io")
	hookRespValidator.AddSchemaFragments(HookResponseSchema)
}

type HookResponse struct {
	IsAllowed bool        `json:"is_allowed"`
	Reason    string      `json:"reason"`
	Data      interface{} `json:"data"`
	Mutations *Mutations  `json:"mutations"`
}

func ParseHookResponse(r io.Reader) (*HookResponse, error) {
	var resp HookResponse
	if err := hookRespValidator.ParseReader("#HookResponse", r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
