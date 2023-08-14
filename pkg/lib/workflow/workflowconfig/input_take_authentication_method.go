package workflowconfig

import (
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputTakeAuthenticationMethod{})
}

var InputTakeAuthenticationMethodSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false,
	"required": ["authentication", "json_pointer"],
	"properties": {
		"authentication": {
			"type": "string",
			"enum": [
				"primary_password",
				"primary_passkey",
				"primary_oob_otp_email",
				"primary_oob_otp_sms",
				"secondary_password",
				"secondary_totp",
				"secondary_oob_otp_email",
				"secondary_oob_otp_sms",
				"recovery_code",
				"device_token"
			]
		},
		"json_pointer": {
			"type": "string",
			"format": "json-pointer"
		}
	}
}
`)

type InputTakeAuthenticationMethod struct {
	Authentication config.WorkflowAuthenticationMethod `json:"authentication,omitempty"`
	JSONPointer    jsonpointer.T                       `json:"json_pointer,omitempty"`
}

func (*InputTakeAuthenticationMethod) Kind() string {
	return "workflowconfig.InputTakeAuthenticationMethod"
}

func (*InputTakeAuthenticationMethod) JSONSchema() *validation.SimpleSchema {
	return InputTakeAuthenticationMethodSchema
}

func (i *InputTakeAuthenticationMethod) GetAuthenticationMethod() config.WorkflowAuthenticationMethod {
	return i.Authentication
}

func (i *InputTakeAuthenticationMethod) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

type inputTakeAuthenticationMethod interface {
	GetAuthenticationMethod() config.WorkflowAuthenticationMethod
	GetJSONPointer() jsonpointer.T
}

var _ inputTakeAuthenticationMethod = &InputTakeAuthenticationMethod{}
