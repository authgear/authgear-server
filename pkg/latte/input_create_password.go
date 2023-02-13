package latte

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputCreatePassword{})
}

var InputCreatePasswordSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"new_password": { "type": "string" }
		},
		"required": ["new_password"]
	}
`)

type InputCreatePassword struct {
	NewPassword string `json:"new_password"`
}

func (*InputCreatePassword) Kind() string {
	return "latte.InputCreatePassword"
}

func (*InputCreatePassword) JSONSchema() *validation.SimpleSchema {
	return InputCreatePasswordSchema
}

func (i *InputCreatePassword) GetNewPassword() string {
	return i.NewPassword
}

type inputCreatePassword interface {
	GetNewPassword() string
}

var _ inputCreatePassword = &InputCreatePassword{}
