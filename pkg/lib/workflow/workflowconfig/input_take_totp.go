package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
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
		"code": { "type": "string" }
	}
}
`)

type InputTakeTOTP struct {
	Code string `json:"code,omitempty"`
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

type inputTakeTOTP interface {
	GetCode() string
}

var _ inputTakeTOTP = &InputTakeTOTP{}
