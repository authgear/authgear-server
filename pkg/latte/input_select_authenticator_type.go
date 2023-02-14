package latte

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputSelectAuthenticatorType{})
}

var InputSelectAuthenticatorTypeSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"authenticator_type": {
				"type": "string",
				"enum": ["password", "oob_otp_email"]
			}
		},
		"required": ["authenticator_type"]
	}
`)

type InputSelectAuthenticatorType struct {
	AuthenticatorType model.AuthenticatorType `json:"authenticator_type"`
}

func (*InputSelectAuthenticatorType) Kind() string {
	return "latte.InputSelectAuthenticatorType"
}

func (*InputSelectAuthenticatorType) JSONSchema() *validation.SimpleSchema {
	return InputSelectAuthenticatorTypeSchema
}

func (i *InputSelectAuthenticatorType) GetAuthenticatorType() model.AuthenticatorType {
	return i.AuthenticatorType
}

type inputSelectAuthenticatorType interface {
	GetAuthenticatorType() model.AuthenticatorType
}

var _ inputSelectAuthenticatorType = &InputSelectAuthenticatorType{}
