package latte

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputResendCode{})
}

var InputResendCodeSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false
	}
`)

type InputResendCode struct{}

func (*InputResendCode) Kind() string {
	return "latte.InputResendCode"
}

func (*InputResendCode) JSONSchema() *validation.SimpleSchema {
	return InputResendCodeSchema
}

func (i *InputResendCode) DoResend() {}

type inputResendCode interface {
	DoResend()
}

var _ inputResendCode = &InputResendCode{}
