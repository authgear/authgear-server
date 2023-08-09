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
		},
		"request_device_token": { "type": "boolean" }
	}
}
`)

type InputTakePassword struct {
	Password           string `json:"password,omitempty"`
	RequestDeviceToken bool   `json:"request_device_token,omitempty"`
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

func (i *InputTakePassword) GetDeviceTokenRequested() bool {
	return i.RequestDeviceToken
}

type inputTakePassword interface {
	GetPassword() string
}

var _ inputTakePassword = &InputTakePassword{}

var _ inputDeviceTokenRequested = &InputTakePassword{}
