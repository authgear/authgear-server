package workflowconfig

import (
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

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
		}
	},
	"required": ["attributes"]
}
`)

type InputFillUserProfileAttribute struct {
	Pointer jsonpointer.T `json:"pointer,omitempty"`
	Value   interface{}   `json:"value,omitempty"`
}

type InputFillUserProfile struct {
	Attributes []InputFillUserProfileAttribute `json:"attributes,omitempty"`
}

func (*InputFillUserProfile) Kind() string {
	return "workflowconfig.InputFillUserProfile"
}

func (*InputFillUserProfile) JSONSchema() *validation.SimpleSchema {
	return InputFillUserProfileSchema
}

func (i *InputFillUserProfile) GetAttributes() []InputFillUserProfileAttribute {
	return i.Attributes
}

type inputFillUserProfile interface {
	GetAttributes() []InputFillUserProfileAttribute
}

var _ inputFillUserProfile = &InputFillUserProfile{}
