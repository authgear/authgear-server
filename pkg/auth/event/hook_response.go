package event

import skyerr "github.com/skygeario/skygear-server/pkg/core/xskyerr"

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
	"type": "object",
	"properties": {
		"is_allowed": { "type": "boolean" },
		"reason": { "type": "string" },
		"data": { "type": "object" },
		"mutations": {
			"type": "object",
			"properties": {
				"is_disabled": { "type": "boolean" },
				"is_verified": { "type": "boolean" },
				"verify_info": { "type": "object" },
				"metadata": { "type": "object" }
			}
		}
	}
}
`

type HookResponse struct {
	IsAllowed bool        `json:"is_allowed"`
	Reason    string      `json:"reason"`
	Data      interface{} `json:"data"`
	Mutations *Mutations  `json:"mutations"`
}

func (resp HookResponse) Validate() error {
	// TODO(error): JSON schema
	if resp.IsAllowed {
		if resp.Reason != "" {
			return skyerr.NewInvalid("reason must not exist")
		}
	} else {
		if resp.Mutations != nil {
			return skyerr.NewInvalid("mutations must not exist")
		}
		if resp.Reason == "" {
			return skyerr.NewInvalid("reason must be provided")
		}
	}
	return nil
}
