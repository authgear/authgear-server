package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputTakePassword{})
}

var InputTakePasswordSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false,
	"required": ["password"],
	"properties": {
		"password": {
			"type": "string"
		}
	}
}
`)

type InputTakePassword struct {
	Password string `json:"password,omitempty"`
}

func (*InputTakePassword) Kind() string {
	return "workflowconfig.InputTakePassword"
}

func (*InputTakePassword) JSONSchema() *validation.SimpleSchema {
	return InputTakePasswordSchema
}

func (i *InputTakePassword) GetPassword() string {
	return i.Password
}

type inputTakePassword interface {
	GetPassword() string
}

var _ inputTakePassword = &InputTakePassword{}
