package workflowconfig

import (
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputTakeTOTP{})
}

var InputTakeTOTPSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false,
	"required": ["code"],
	"properties": {
		"code": { "type": "string" },
		"request_device_token": { "type": "boolean" }
	}
}
`)

type InputTakeTOTP struct {
	Code               string `json:"code,omitempty"`
	RequestDeviceToken bool   `json:"request_device_token,omitempty"`
}

func (*InputTakeTOTP) Kind() string {
	return "workflowconfig.InputTakeTOTP"
}

func (*InputTakeTOTP) JSONSchema() *validation.SimpleSchema {
	return InputTakeTOTPSchema
}

func (i *InputTakeTOTP) GetCode() string {
	return i.Code
}

func (i *InputTakeTOTP) GetDeviceTokenRequested() bool {
	return i.RequestDeviceToken
}

type inputTakeTOTP interface {
	GetCode() string
}

var _ inputTakeTOTP = &InputTakeTOTP{}

var _ inputDeviceTokenRequested = &InputTakeTOTP{}
