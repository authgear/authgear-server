package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
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
		}
	}
}
`)

type InputTakeRecoveryCode struct {
	RecoveryCode string `json:"recovery_code,omitempty"`
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

type inputTakeRecoveryCode interface {
	GetRecoveryCode() string
}

var _ inputTakeRecoveryCode = &InputTakeRecoveryCode{}
