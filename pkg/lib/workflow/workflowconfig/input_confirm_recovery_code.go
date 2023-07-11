package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputConfirmRecoveryCode{})
}

var InputConfirmRecoveryCodeSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false
}
`)

type InputConfirmRecoveryCode struct{}

func (*InputConfirmRecoveryCode) Kind() string {
	return "workflowconfig.InputConfirmRecoveryCode"
}

func (*InputConfirmRecoveryCode) JSONSchema() *validation.SimpleSchema {
	return InputConfirmRecoveryCodeSchema
}

func (*InputConfirmRecoveryCode) ConfirmRecoveryCode() {}

type inputConfirmRecoveryCode interface {
	ConfirmRecoveryCode()
}

var _ inputConfirmRecoveryCode = &InputConfirmRecoveryCode{}
