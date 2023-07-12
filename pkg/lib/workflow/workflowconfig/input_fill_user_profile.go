package workflowconfig

import (
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/lib/authn/attrs"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputFillUserProfile{})
}

var InputFillUserProfileSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"attributes": {
			"type": "array",
			"items": {
				"type": "object",
				"required": ["pointer", "value"],
				"properties": {
					"pointer": {
						"type": "string",
						"format": "json-pointer"
					},
					"value": {}
				}
			}
		},
		"json_pointer": {
			"type": "string",
			"format": "json-pointer"
		}
	},
	"required": ["attributes", "json_pointer"]
}
`)

type InputFillUserProfile struct {
	Attributes  []attrs.T     `json:"attributes,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
}

func (*InputFillUserProfile) Kind() string {
	return "workflowconfig.InputFillUserProfile"
}

func (*InputFillUserProfile) JSONSchema() *validation.SimpleSchema {
	return InputFillUserProfileSchema
}

func (i *InputFillUserProfile) GetAttributes() []attrs.T {
	return i.Attributes
}

func (i *InputFillUserProfile) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

type inputFillUserProfile interface {
	GetAttributes() []attrs.T
	GetJSONPointer() jsonpointer.T
}

var _ inputFillUserProfile = &InputFillUserProfile{}
