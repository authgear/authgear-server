package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputTakeAuthenticatorID{})
}

var InputTakeAuthenticatorIDSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"authenticator_id": { "type": "string" }
	},
	"required": ["authenticator_id"]
}
`)

type InputTakeAuthenticatorID struct {
	AuthenticatorID string `json:"authenticator_id"`
}

func (*InputTakeAuthenticatorID) Kind() string {
	return "workflowconfig.InputTakeAuthenticatorID"
}

func (*InputTakeAuthenticatorID) JSONSchema() *validation.SimpleSchema {
	return InputTakeAuthenticatorIDSchema
}

func (i *InputTakeAuthenticatorID) GetAuthenticatorID() string {
	return i.AuthenticatorID
}

type inputTakeAuthenticatorID interface {
	GetAuthenticatorID() string
}

var _ inputTakeAuthenticatorID = &InputTakeAuthenticatorID{}
