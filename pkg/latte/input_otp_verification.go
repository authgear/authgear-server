package latte

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputOTPVerification{})
}

var InputOTPVerificationSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"code": { "type": "string" }
		},
		"required": ["code"]
	}
`)

type InputOTPVerification struct {
	Code string `json:"code"`
}

func (*InputOTPVerification) Kind() string {
	return "latte.InputOTPVerification"
}

func (*InputOTPVerification) JSONSchema() *validation.SimpleSchema {
	return InputOTPVerificationSchema
}

func (i *InputOTPVerification) GetOTPVerification() string {
	return i.Code
}

type inputOTPVerification interface {
	GetOTPVerification() string
}

var _ inputOTPVerification = &InputOTPVerification{}
