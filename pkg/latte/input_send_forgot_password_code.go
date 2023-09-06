package latte

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputSendForgotPasswordCode{})
}

var InputSendForgotPasswordCodeSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false
	}
`)

type InputSendForgotPasswordCode struct{}

func (*InputSendForgotPasswordCode) Kind() string {
	return "latte.InputSendForgotPasswordCode"
}

func (*InputSendForgotPasswordCode) JSONSchema() *validation.SimpleSchema {
	return InputSendForgotPasswordCodeSchema
}

func (i *InputSendForgotPasswordCode) DoSend() {}

type inputSendForgotPasswordCode interface {
	DoSend()
}

var _ inputSendForgotPasswordCode = &InputSendForgotPasswordCode{}
