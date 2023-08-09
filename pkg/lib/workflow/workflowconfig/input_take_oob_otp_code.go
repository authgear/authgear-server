package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputTakeOOBOTPCode{})
}

var InputTakeOOBOTPCodeSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"code": { "type": "string" },
		"request_device_token": { "type": "boolean" }
	},
	"required": ["code"]
}
`)

type InputTakeOOBOTPCode struct {
	Code               string `json:"code,omitempty"`
	RequestDeviceToken bool   `json:"request_device_token,omitempty"`
}

func (*InputTakeOOBOTPCode) Kind() string {
	return "workflowconfig.InputTakeOOBOTPCode"
}

func (*InputTakeOOBOTPCode) JSONSchema() *validation.SimpleSchema {
	return InputTakeOOBOTPCodeSchema
}

func (i *InputTakeOOBOTPCode) GetCode() string {
	return i.Code
}

func (i *InputTakeOOBOTPCode) GetDeviceTokenRequested() bool {
	return i.RequestDeviceToken
}

type inputTakeOOBOTPCode interface {
	GetCode() string
}

var _ inputTakeOOBOTPCode = &InputTakeOOBOTPCode{}

var _ inputDeviceTokenRequested = &InputTakeOOBOTPCode{}
