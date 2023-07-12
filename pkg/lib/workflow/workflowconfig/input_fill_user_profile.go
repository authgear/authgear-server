package workflowconfig

import (
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
		}
	},
	"required": ["attributes"]
}
`)

type InputFillUserProfile struct {
	Attributes []attrs.T `json:"attributes,omitempty"`
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

type inputFillUserProfile interface {
	GetAttributes() []attrs.T
}

var _ inputFillUserProfile = &InputFillUserProfile{}
