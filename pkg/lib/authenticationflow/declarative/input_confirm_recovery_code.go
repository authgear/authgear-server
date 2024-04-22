package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
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

type InputConfirmRecoveryCode struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
}

var _ authflow.InputSchema = &InputConfirmRecoveryCode{}
var _ authflow.Input = &InputConfirmRecoveryCode{}
var _ inputConfirmRecoveryCode = &InputConfirmRecoveryCode{}

func (i *InputConfirmRecoveryCode) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputConfirmRecoveryCode) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

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
