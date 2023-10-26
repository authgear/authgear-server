package authenticationflow

import (
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/util/validation"
)

var InputTakeAccountRecoveryCodeSchemaBuilder validation.SchemaBuilder

func init() {
	InputTakeAccountRecoveryCodeSchemaBuilder = validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("account_recovery_code")

	InputTakeAccountRecoveryCodeSchemaBuilder.Properties().Property(
		"account_recovery_code",
		validation.SchemaBuilder{}.Type(validation.TypeString),
	)
}

type InputTakeAccountRecoveryCode struct {
	AccountRecoveryCode string `json:"account_recovery_code"`
}

func MakeInputTakeAccountRecoveryCode(rawMessage json.RawMessage) (*InputTakeAccountRecoveryCode, bool) {
	var input InputTakeAccountRecoveryCode
	err := InputTakeAccountRecoveryCodeSchemaBuilder.ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, false
	}
	return &input, true
}
