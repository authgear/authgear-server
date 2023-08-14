package workflowconfig

import (
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputTakeOOBOTPTarget{})
}

var InputTakeOOBOTPTargetSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"target": { "type": "string" }
		},
		"required": ["target"]
	}
`)

type InputTakeOOBOTPTarget struct {
	Target string `json:"target"`
}

func (*InputTakeOOBOTPTarget) Kind() string {
	return "workflowconfig.InputTakeOOBOTPTarget"
}

func (*InputTakeOOBOTPTarget) JSONSchema() *validation.SimpleSchema {
	return InputTakeOOBOTPTargetSchema
}

func (i *InputTakeOOBOTPTarget) GetTarget() string {
	return i.Target
}

type inputTakeOOBOTPTarget interface {
	GetTarget() string
}

var _ inputTakeOOBOTPTarget = &InputTakeOOBOTPTarget{}
