package latte

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputTakeLoginLinkCode{})
}

var InputTakeLoginLinkCodeSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"code": { "type": "string" }
		},
		"required": ["code"]
	}
`)

type InputTakeLoginLinkCode struct {
	Code string `json:"code"`
}

func (*InputTakeLoginLinkCode) Kind() string {
	return "latte.InputTakeLoginLinkCode"
}

func (*InputTakeLoginLinkCode) JSONSchema() *validation.SimpleSchema {
	return InputTakeLoginLinkCodeSchema
}

func (i *InputTakeLoginLinkCode) GetCode() string {
	return i.Code
}

type inputTakeLoginLinkCode interface {
	GetCode() string
}

var _ inputTakeLoginLinkCode = &InputTakeLoginLinkCode{}
