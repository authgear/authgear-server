package workflowconfig

import (
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/lib/config"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputTakeIdentificationMethod{})
}

var InputTakeIdentificationMethodSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false,
	"required": ["identification", "json_pointer"],
	"properties": {
		"identification": {
			"type": "string",
			"enum": [
				"email",
				"phone",
				"username",
				"oauth",
				"passkey",
				"siwe"
			]
		},
		"json_pointer": {
			"type": "string",
			"format": "json-pointer"
		}
	}
}
`)

type InputTakeIdentificationMethod struct {
	Identification config.WorkflowIdentificationMethod `json:"identification,omitempty"`
	JSONPointer    jsonpointer.T                       `json:"json_pointer,omitempty"`
}

func (*InputTakeIdentificationMethod) Kind() string {
	return "workflowconfig.InputTakeIdentificationMethod"
}

func (*InputTakeIdentificationMethod) JSONSchema() *validation.SimpleSchema {
	return InputTakeIdentificationMethodSchema
}

func (i *InputTakeIdentificationMethod) GetIdentificationMethod() config.WorkflowIdentificationMethod {
	return i.Identification
}

func (i *InputTakeIdentificationMethod) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

type inputTakeIdentificationMethod interface {
	GetIdentificationMethod() config.WorkflowIdentificationMethod
	GetJSONPointer() jsonpointer.T
}

var _ inputTakeIdentificationMethod = &InputTakeIdentificationMethod{}
