package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputTakeOOBOTPChannel{})
}

var InputTakeOOBOTPChannelSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"channel": {
			"type": "string",
			"enum": [
				"sms",
				"email",
				"whatsapp"
			]
		}
	},
	"required": ["channel"]
}
`)

type InputTakeOOBOTPChannel struct {
	Channel model.AuthenticatorOOBChannel `json:"channel,omitempty"`
}

func (*InputTakeOOBOTPChannel) Kind() string {
	return "workflowconfig.InputTakeOOBOTPChannel"
}

func (*InputTakeOOBOTPChannel) JSONSchema() *validation.SimpleSchema {
	return InputTakeOOBOTPChannelSchema
}

func (i *InputTakeOOBOTPChannel) GetChannel() model.AuthenticatorOOBChannel {
	return i.Channel
}

type inputTakeOOBOTPChannel interface {
	GetChannel() model.AuthenticatorOOBChannel
}

var _ inputTakeOOBOTPChannel = &InputTakeOOBOTPChannel{}
