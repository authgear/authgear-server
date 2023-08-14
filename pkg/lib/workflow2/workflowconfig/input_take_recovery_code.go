package workflowconfig

import (
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputTakeRecoveryCode{})
}

var InputTakeRecoveryCodeSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false,
	"required": ["recovery_code"],
	"properties": {
		"recovery_code": {
			"type": "string"
		},
		"request_device_token": { "type": "boolean" }
	}
}
`)

type InputTakeRecoveryCode struct {
	RecoveryCode       string `json:"recovery_code,omitempty"`
	RequestDeviceToken bool   `json:"request_device_token,omitempty"`
}

func (*InputTakeRecoveryCode) Kind() string {
	return "workflowconfig.InputTakeRecoveryCode"
}

func (*InputTakeRecoveryCode) JSONSchema() *validation.SimpleSchema {
	return InputTakeRecoveryCodeSchema
}

func (i *InputTakeRecoveryCode) GetRecoveryCode() string {
	return i.RecoveryCode
}

func (i *InputTakeRecoveryCode) GetDeviceTokenRequested() bool {
	return i.RequestDeviceToken
}

type inputTakeRecoveryCode interface {
	GetRecoveryCode() string
}

var _ inputTakeRecoveryCode = &InputTakeRecoveryCode{}

var _ inputDeviceTokenRequested = &InputTakeRecoveryCode{}
