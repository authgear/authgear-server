package latte

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type VerificationMethod string

const (
	VerificationMethodEmail    VerificationMethod = "email"
	VerificationMethodPhoneSMS VerificationMethod = "sms"
)

func init() {
	workflow.RegisterPublicInput(&InputTriggerVerification{})
}

var InputTriggerVerificationSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"method": {
				"type": "string",
				"enum": ["email", "sms"]
			},
			"claim_name": { "type": "string" },
			"claim_value": { "type": "string" }
		},
		"required": ["method", "claim_name", "claim_value"]
	}
`)

type InputTriggerVerification struct {
	Method     VerificationMethod `json:"method"`
	ClaimName  string             `json:"claim_name"`
	ClaimValue string             `json:"claim_value"`
}

func (*InputTriggerVerification) Kind() string {
	return "latte.InputTriggerVerification"
}

func (*InputTriggerVerification) JSONSchema() *validation.SimpleSchema {
	return InputTriggerVerificationSchema
}

func (i *InputTriggerVerification) VerificationMethod() VerificationMethod {
	return i.Method
}

func (i *InputTriggerVerification) ClaimToVerify() (name string, value string) {
	return i.ClaimName, i.ClaimValue
}

type inputTriggerVerification interface {
	VerificationMethod() VerificationMethod
	ClaimToVerify() (name string, value string)
}

var _ inputTriggerVerification = &InputTriggerVerification{}
