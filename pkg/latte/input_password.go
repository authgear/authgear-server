package latte

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputPassword{})
}

var InputPasswordSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"password": { "type": "string" }
		},
		"required": ["password"]
	}
`)

type InputPassword struct {
	Password string `json:"password"`
}

func (*InputPassword) Kind() string {
	return "latte.InputPassword"
}

func (*InputPassword) JSONSchema() *validation.SimpleSchema {
	return InputPasswordSchema
}

func (i *InputPassword) GetPassword() string {
	return i.Password
}

type inputPassword interface {
	GetPassword() string
}

var _ inputPassword = &InputPassword{}
