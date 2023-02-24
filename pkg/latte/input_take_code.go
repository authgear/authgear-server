package latte

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputTakeCode{})
}

var InputTakeCodeSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"code": { "type": "string" }
		},
		"required": ["code"]
	}
`)

type InputTakeCode struct {
	Code string `json:"code"`
}

func (*InputTakeCode) Kind() string {
	return "latte.InputTakeCode"
}

func (*InputTakeCode) JSONSchema() *validation.SimpleSchema {
	return InputTakeCodeSchema
}

func (i *InputTakeCode) GetCode() string {
	return i.Code
}

type inputTakeCode interface {
	GetCode() string
}

var _ inputTakeCode = &InputTakeCode{}
