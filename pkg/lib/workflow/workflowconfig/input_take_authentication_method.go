package workflowconfig

import (
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
	"required": ["authentication"],
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
				"secondary_oob_otp_sms"
			]
		}
	}
}
`)

type InputTakeAuthenticationMethod struct {
	Authentication config.WorkflowAuthenticationMethod `json:"authentication,omitempty"`
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

type inputTakeAuthenticationMethod interface {
	GetAuthenticationMethod() config.WorkflowAuthenticationMethod
}

var _ inputTakeAuthenticationMethod = &InputTakeAuthenticationMethod{}
