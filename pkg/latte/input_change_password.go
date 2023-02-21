package latte

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputChangePassword{})
}

var InputChangePasswordSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"old_password": { "type": "string" },
			"new_password": { "type": "string" }
		},
		"required": ["old_password", "new_password"]
	}
`)

type InputChangePassword struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (*InputChangePassword) Kind() string {
	return "latte.InputChangePassword"
}

func (*InputChangePassword) JSONSchema() *validation.SimpleSchema {
	return InputChangePasswordSchema
}

func (i *InputChangePassword) GetOldPassword() string {
	return i.OldPassword
}

func (i *InputChangePassword) GetNewPassword() string {
	return i.NewPassword
}

type inputChangePassword interface {
	GetOldPassword() string
	GetNewPassword() string
}

var _ inputChangePassword = &InputChangePassword{}
