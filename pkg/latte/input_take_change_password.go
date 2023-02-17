package latte

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputTakeChangePassword{})
}

var InputTakeChangePasswordSchema = validation.NewSimpleSchema(`
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

type InputTakeChangePassword struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (*InputTakeChangePassword) Kind() string {
	return "latte.InputTakeChangePassword"
}

func (*InputTakeChangePassword) JSONSchema() *validation.SimpleSchema {
	return InputTakeChangePasswordSchema
}

func (i *InputTakeChangePassword) GetOldPassword() string {
	return i.OldPassword
}

func (i *InputTakeChangePassword) GetNewPassword() string {
	return i.NewPassword
}

type inputTakeChangePassword interface {
	GetOldPassword() string
	GetNewPassword() string
}

var _ inputTakeChangePassword = &InputTakeChangePassword{}
