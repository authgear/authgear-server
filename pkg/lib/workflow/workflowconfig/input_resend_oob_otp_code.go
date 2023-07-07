package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputResendOOBOTPCode{})
}

var InputResendOOBOTPCodeSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false
}
`)

type InputResendOOBOTPCode struct{}

func (*InputResendOOBOTPCode) Kind() string {
	return "workflowconfig.InputResendOOBOTPCode"
}

func (*InputResendOOBOTPCode) JSONSchema() *validation.SimpleSchema {
	return InputResendOOBOTPCodeSchema
}

func (*InputResendOOBOTPCode) ResendCode() {}

type inputResendOOBOTPCode interface {
	ResendCode()
}

var _ inputResendOOBOTPCode = &InputResendOOBOTPCode{}
