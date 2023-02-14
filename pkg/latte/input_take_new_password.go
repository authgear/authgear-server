package latte

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputTakeNewPassword{})
}

var InputTakeNewPasswordSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"new_password": { "type": "string" }
		},
		"required": ["new_password"]
	}
`)

type InputTakeNewPassword struct {
	NewPassword string `json:"new_password"`
}

func (*InputTakeNewPassword) Kind() string {
	return "latte.InputTakeNewPassword"
}

func (*InputTakeNewPassword) JSONSchema() *validation.SimpleSchema {
	return InputTakeNewPasswordSchema
}

func (i *InputTakeNewPassword) GetNewPassword() string {
	return i.NewPassword
}

type inputTakeNewPassword interface {
	GetNewPassword() string
}

var _ inputTakeNewPassword = &InputTakeNewPassword{}
