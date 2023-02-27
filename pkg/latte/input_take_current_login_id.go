package latte

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputTakeCurrentLoginID{})
}

var InputTakeCurrentLoginIDSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"login_id": { "type": "string" }
		},
		"required": ["login_id"]
	}
`)

type InputTakeCurrentLoginID struct {
	CurrentLoginID string `json:"login_id"`
}

func (*InputTakeCurrentLoginID) Kind() string {
	return "latte.InputTakeCurrentLoginID"
}

func (*InputTakeCurrentLoginID) JSONSchema() *validation.SimpleSchema {
	return InputTakeCurrentLoginIDSchema
}

func (i *InputTakeCurrentLoginID) GetCurrentLoginID() string {
	return i.CurrentLoginID
}

type inputTakeCurrentLoginID interface {
	GetCurrentLoginID() string
}

var _ inputTakeCurrentLoginID = &InputTakeCurrentLoginID{}
