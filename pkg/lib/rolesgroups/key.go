package rolesgroups

import (
	"context"

	"github.com/authgear/authgear-server/pkg/util/validation"
)

var KeySchema = validation.NewSimpleSchema(`
	{
		"type": "string",
		"minLength": 1,
		"maxLength": 40,
		"pattern": "^[a-zA-Z_][a-zA-Z0-9:_]*$",
		"format": "x_role_group_key"
	}
`)

func ValidateKey(ctx context.Context, key string) error {
	return KeySchema.Validator().ValidateValue(ctx, key)
}
