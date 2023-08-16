package workflowconfig

import (
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputTakeAuthenticationCandidateIndex{})
}

var InputTakeAuthenticationCandidateIndexSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"index": { "type": "integer", "min": 0 }
	},
	"required": ["index"]
}
`)

type InputTakeAuthenticationCandidateIndex struct {
	Index int `json:"index"`
}

func (*InputTakeAuthenticationCandidateIndex) Kind() string {
	return "workflowconfig.InputTakeAuthenticationCandidateIndex"
}

func (*InputTakeAuthenticationCandidateIndex) JSONSchema() *validation.SimpleSchema {
	return InputTakeAuthenticationCandidateIndexSchema
}

func (i *InputTakeAuthenticationCandidateIndex) GetIndex() int {
	return i.Index
}

type inputTakeAuthenticationCandidateIndex interface {
	GetIndex() int
}

var _ inputTakeAuthenticationCandidateIndex = &InputTakeAuthenticationCandidateIndex{}
