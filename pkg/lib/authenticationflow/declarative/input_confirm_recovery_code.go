package declarative

import (
	"encoding/json"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var InputConfirmRecoveryCodeSchemaBuilder validation.SchemaBuilder

func init() {
	InputConfirmRecoveryCodeSchemaBuilder = validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("confirm_recovery_code")

	InputConfirmRecoveryCodeSchemaBuilder.Properties().Property(
		"confirm_recovery_code",
		validation.SchemaBuilder{}.
			Type(validation.TypeBoolean).
			Const(true),
	)
}

type InputConfirmRecoveryCode struct{}

var _ authflow.InputSchema = &InputConfirmRecoveryCode{}
var _ authflow.Input = &InputConfirmRecoveryCode{}
var _ inputConfirmRecoveryCode = &InputConfirmRecoveryCode{}

func (*InputConfirmRecoveryCode) SchemaBuilder() validation.SchemaBuilder {
	return InputConfirmRecoveryCodeSchemaBuilder
}

func (i *InputConfirmRecoveryCode) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputConfirmRecoveryCode
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

func (*InputConfirmRecoveryCode) Input() {}

func (*InputConfirmRecoveryCode) ConfirmRecoveryCode() {}
