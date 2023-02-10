package latte

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputOTPVerificationResend{})
}

var InputOTPVerificationResendSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false
	}
`)

type InputOTPVerificationResend struct {
}

func (*InputOTPVerificationResend) Kind() string {
	return "latte.InputOTPVerificationResend"
}

func (*InputOTPVerificationResend) JSONSchema() *validation.SimpleSchema {
	return InputOTPVerificationResendSchema
}

func (i *InputOTPVerificationResend) ResendOTPVerification() {}

type inputOTPVerificationResend interface {
	ResendOTPVerification()
}

var _ inputOTPVerificationResend = &InputOTPVerificationResend{}
