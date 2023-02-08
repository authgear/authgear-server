package latte

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputTakeLoginID{})
}

var InputTakeLoginIDSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"login_id": { "type": "string" }
		},
		"required": ["login_id"]
	}
`)

type InputTakeLoginID struct {
	LoginID string `json:"login_id"`
}

func (*InputTakeLoginID) Kind() string {
	return "latte.InputTakeLoginID"
}

func (*InputTakeLoginID) JSONSchema() *validation.SimpleSchema {
	return InputTakeLoginIDSchema
}

func (i *InputTakeLoginID) GetLoginID() string {
	return i.LoginID
}

type inputTakeLoginID interface {
	GetLoginID() string
}

var _ inputTakeLoginID = &InputTakeLoginID{}
