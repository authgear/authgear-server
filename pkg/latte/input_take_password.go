package latte

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
		"properties": {
			"password": { "type": "string" }
		},
		"required": ["password"]
	}
`)

type InputTakePassword struct {
	Password string `json:"password"`
}

func (*InputTakePassword) Kind() string {
	return "latte.InputTakePassword"
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
