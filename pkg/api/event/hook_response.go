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
	"type": "object",
	"properties": {
		"is_allowed": { "type": "boolean" },
		"title": { "type": "string" },
		"reason": { "type": "string" }
	},
	"required": ["is_allowed"]
}
`)

type HookResponse struct {
	IsAllowed bool   `json:"is_allowed"`
	Title     string `json:"title"`
	Reason    string `json:"reason"`
}

func ParseHookResponse(r io.Reader) (*HookResponse, error) {
	var resp HookResponse
	if err := HookResponseSchema.Validator().Parse(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
