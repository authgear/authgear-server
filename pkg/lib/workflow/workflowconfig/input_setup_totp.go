package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputSetupTOTP{})
}

var InputSetupTOTPSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false,
	"required": ["code", "display_name"],
	"properties": {
		"code": { "type": "string" },
		"display_name": { "type": "string" }
	}
}
`)

type InputSetupTOTP struct {
	Code        string `json:"code,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
}

func (*InputSetupTOTP) Kind() string {
	return "workflowconfig.InputSetupTOTP"
}

func (*InputSetupTOTP) JSONSchema() *validation.SimpleSchema {
	return InputSetupTOTPSchema
}

func (i *InputSetupTOTP) GetCode() string {
	return i.Code
}

func (i *InputSetupTOTP) GetDisplayName() string {
	return i.DisplayName
}

type inputSetupTOTP interface {
	GetCode() string
	GetDisplayName() string
}

var _ inputSetupTOTP = &InputSetupTOTP{}
