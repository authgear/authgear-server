package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
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

type InputSchemaTakeAccountRecoveryCode struct {
	JSONPointer jsonpointer.T
}

var _ authflow.InputSchema = &InputSchemaTakeAccountRecoveryCode{}

func (i *InputSchemaTakeAccountRecoveryCode) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (*InputSchemaTakeAccountRecoveryCode) SchemaBuilder() validation.SchemaBuilder {
	return InputTakeAccountRecoveryCodeSchemaBuilder
}

func (i *InputSchemaTakeAccountRecoveryCode) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputTakeAccountRecoveryCode
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputTakeAccountRecoveryCode struct {
	AccountRecoveryCode string `json:"account_recovery_code"`
}

var _ authflow.Input = &InputTakeAccountRecoveryCode{}
var _ inputTakeAccountRecoveryCode = &InputTakeAccountRecoveryCode{}

func (*InputTakeAccountRecoveryCode) Input() {}

func (i *InputTakeAccountRecoveryCode) GetAccountRecoveryCode() string {
	return i.AccountRecoveryCode
}
