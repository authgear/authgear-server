package latte

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputTakeProofOfPhoneNumberVerification{})
}

var InputTakeProofOfPhoneNumberVerificationSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"proof_of_phone_number_verification": { "type": "string" }
		},
		"required": ["proof_of_phone_number_verification"]
	}
`)

type InputTakeProofOfPhoneNumberVerification struct {
	ProofOfPhoneNumberVerification string `json:"proof_of_phone_number_verification"`
}

func (*InputTakeProofOfPhoneNumberVerification) Kind() string {
	return "latte.InputTakeProofOfPhoneNumberVerification"
}

func (*InputTakeProofOfPhoneNumberVerification) JSONSchema() *validation.SimpleSchema {
	return InputTakeProofOfPhoneNumberVerificationSchema
}

func (i *InputTakeProofOfPhoneNumberVerification) GetProofOfPhoneNumberVerification() string {
	return i.ProofOfPhoneNumberVerification
}

type inputTakeProofOfPhoneNumberVerification interface {
	GetProofOfPhoneNumberVerification() string
}

var _ inputTakeProofOfPhoneNumberVerification = &InputTakeProofOfPhoneNumberVerification{}
