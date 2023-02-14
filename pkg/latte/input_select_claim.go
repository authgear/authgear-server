package latte

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputSelectClaim{})
}

var InputSelectClaimSchema = validation.NewSimpleSchema(`
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

type InputSelectClaim struct {
	ClaimName  string `json:"claim_name"`
	ClaimValue string `json:"claim_value"`
}

func (*InputSelectClaim) Kind() string {
	return "latte.InputSelectClaim"
}

func (*InputSelectClaim) JSONSchema() *validation.SimpleSchema {
	return InputSelectClaimSchema
}

func (i *InputSelectClaim) NameValue() (name string, value string) {
	return i.ClaimName, i.ClaimValue
}

type inputSelectClaim interface {
	NameValue() (name string, value string)
}

var _ inputSelectClaim = &InputSelectClaim{}
