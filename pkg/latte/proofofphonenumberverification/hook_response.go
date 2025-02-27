package proofofphonenumberverification

import (
	"context"
	"io"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var HookResponseSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"identity": {
			"type": "object",
			"additionalProperties": false,
			"properties": {
				"type" : {
					"type": "string",
					"enum" : ["login_id"]
				},
				"login_id": {
					"type": "object",
					"properties": {
						"key": { "type": "string" },
						"type": { "type": "string" },
						"value": { "type": "string" }
					},
					"required": ["key", "type", "value"]
				}
			}
		}
	},
	"required": ["identity"]
}
`)

type HookResponse struct {
	Identity *identity.Spec `json:"identity"`
}

func ParseHookResponse(ctx context.Context, r io.Reader) (*HookResponse, error) {
	var resp HookResponse
	if err := HookResponseSchema.Validator().Parse(ctx, r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
