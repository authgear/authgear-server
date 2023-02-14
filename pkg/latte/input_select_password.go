package latte

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputSelectPassword{})
}

var InputSelectPasswordSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false
	}
`)

type InputSelectPassword struct{}

func (*InputSelectPassword) Kind() string {
	return "latte.InputSelectPassword"
}

func (*InputSelectPassword) JSONSchema() *validation.SimpleSchema {
	return InputSelectPasswordSchema
}

func (i *InputSelectPassword) SelectPassword() {}

type inputSelectPassword interface {
	SelectPassword()
}

var _ inputSelectPassword = &InputSelectPassword{}
