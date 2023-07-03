package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputIdentify{})
}

var InputIdentifySchema = validation.NewSimpleSchema(`
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

type InputIdentify struct {
	Identification config.WorkflowIdentificationMethod `json:"identification,omitempty"`
}

func (*InputIdentify) Kind() string {
	return "workflowconfig.InputIdentify"
}

func (*InputIdentify) JSONSchema() *validation.SimpleSchema {
	return InputIdentifySchema
}

func (i *InputIdentify) GetIdentificationMethod() config.WorkflowIdentificationMethod {
	return i.Identification
}

type inputIdentify interface {
	GetIdentificationMethod() config.WorkflowIdentificationMethod
}

var _ inputIdentify = &InputIdentify{}
