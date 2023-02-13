package latte

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputTriggerVerification{})
}

var InputTriggerVerificationSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"claim_name": { "type": "string" },
			"claim_value": { "type": "string" }
		},
		"required": ["claim_name", "claim_value"]
	}
`)

type InputTriggerVerification struct {
	ClaimName  string `json:"claim_name"`
	ClaimValue string `json:"claim_value"`
}

func (*InputTriggerVerification) Kind() string {
	return "latte.InputTriggerVerification"
}

func (*InputTriggerVerification) JSONSchema() *validation.SimpleSchema {
	return InputTriggerVerificationSchema
}

func (i *InputTriggerVerification) ClaimToVerify() (name string, value string) {
	return i.ClaimName, i.ClaimValue
}

type inputTriggerVerification interface {
	ClaimToVerify() (name string, value string)
}

var _ inputTriggerVerification = &InputTriggerVerification{}
