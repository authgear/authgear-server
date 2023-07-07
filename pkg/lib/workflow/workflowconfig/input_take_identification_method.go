package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputTakeIdentificationMethod{})
}

var InputTakeIdentificationMethodSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false,
	"required": ["identification"],
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
		}
	}
}
`)

type InputTakeIdentificationMethod struct {
	Identification config.WorkflowIdentificationMethod `json:"identification,omitempty"`
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

type inputTakeIdentificationMethod interface {
	GetIdentificationMethod() config.WorkflowIdentificationMethod
}

var _ inputTakeIdentificationMethod = &InputTakeIdentificationMethod{}
