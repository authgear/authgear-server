package latte

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputTakeForgotPasswordChannel{})
}

var InputTakeForgotPasswordChannelSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"channel": {
				"type": "string",
				"enum": ["email", "sms"]
			}
		},
		"required": ["channel"]
	}
`)

type InputTakeForgotPasswordChannel struct {
	Channel ForgotPasswordChannel `json:"channel"`
}

func (*InputTakeForgotPasswordChannel) Kind() string {
	return "latte.InputTakeForgotPasswordChannel"
}

func (*InputTakeForgotPasswordChannel) JSONSchema() *validation.SimpleSchema {
	return InputTakeForgotPasswordChannelSchema
}

func (i *InputTakeForgotPasswordChannel) GetForgotPasswordChannel() ForgotPasswordChannel {
	return i.Channel
}

type inputTakeForgotPasswordChannel interface {
	GetForgotPasswordChannel() ForgotPasswordChannel
}

var _ inputTakeForgotPasswordChannel = &InputTakeForgotPasswordChannel{}
