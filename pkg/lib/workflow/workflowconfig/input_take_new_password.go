package workflowconfig

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
	"required": ["new_password"],
	"properties": {
		"new_password": {
			"type": "string"
		}
	}
}
`)

type InputTakeNewPassword struct {
	NewPassword string `json:"new_password,omitempty"`
}

func (*InputTakeNewPassword) Kind() string {
	return "workflowconfig.InputTakeNewPassword"
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
