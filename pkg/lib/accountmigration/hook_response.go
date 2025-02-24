package accountmigration

import (
	"context"
	"io"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var HookResponseSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"identities": {
			"type": "array",
			"minItems": 1,
			"items": {
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
				},
				"required": ["type", "login_id"]
			}
		},
		"authenticators": {
			"type": "array",
			"items": {
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"type" : {
						"type": "string",
						"enum" : ["oob_otp_email", "oob_otp_sms"]
					},
					"oobotp": {
						"email": { "type": "string" },
						"phone": { "type": "string" }
					}
				},
				"allOf": [
					{
						"if": { "properties": { "type": { "const": "oob_otp_email" } } },
						"then": {
							"properties": {
								"oobotp": {
									"required": ["email"]
								}
							},
							"required": ["oobotp"]
						}
					},
					{
						"if": { "properties": { "type": { "const": "oob_otp_sms" } } },
						"then": {
							"properties": {
								"oobotp": {
									"required": ["phone"]
								}
							},
							"required": ["oobotp"]
						}
					}
				]
			}
		}
	},
	"required": ["identities"]
}
`)

type HookResponse struct {
	Identities     []*identity.MigrateSpec
	Authenticators []*authenticator.MigrateSpec
}

func ParseHookResponse(ctx context.Context, r io.Reader) (*HookResponse, error) {
	var resp HookResponse
	if err := HookResponseSchema.Validator().Parse(ctx, r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
