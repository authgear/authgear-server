package latte

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputTakeOOBOTPCode{})
}

var InputTakeOOBOTPCodeSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"code": { "type": "string" }
		},
		"required": ["code"]
	}
`)

type InputTakeOOBOTPCode struct {
	Code string `json:"code"`
}

func (*InputTakeOOBOTPCode) Kind() string {
	return "latte.InputTakeOOBOTPCode"
}

func (*InputTakeOOBOTPCode) JSONSchema() *validation.SimpleSchema {
	return InputTakeOOBOTPCodeSchema
}

func (i *InputTakeOOBOTPCode) GetCode() string {
	return i.Code
}

type inputTakeOOBOTPCode interface {
	GetCode() string
}

var _ inputTakeOOBOTPCode = &InputTakeOOBOTPCode{}
