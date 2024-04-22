package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var InputConfirmTerminateOtherSessionsSchemaBuilder validation.SchemaBuilder

func init() {
	InputConfirmTerminateOtherSessionsSchemaBuilder = validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("confirm_terminate_other_sessions")

	InputConfirmTerminateOtherSessionsSchemaBuilder.Properties().Property(
		"confirm_terminate_other_sessions",
		validation.SchemaBuilder{}.
			Type(validation.TypeBoolean).
			Const(true),
	)

}

type InputConfirmTerminateOtherSessions struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
}

var _ authflow.InputSchema = &InputConfirmTerminateOtherSessions{}
var _ authflow.Input = &InputConfirmTerminateOtherSessions{}
var _ inputConfirmTerminateOtherSessions = &InputConfirmTerminateOtherSessions{}

func (i *InputConfirmTerminateOtherSessions) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputConfirmTerminateOtherSessions) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (*InputConfirmTerminateOtherSessions) SchemaBuilder() validation.SchemaBuilder {
	return InputConfirmTerminateOtherSessionsSchemaBuilder
}

func (i *InputConfirmTerminateOtherSessions) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputConfirmTerminateOtherSessions
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

func (*InputConfirmTerminateOtherSessions) Input() {}

func (*InputConfirmTerminateOtherSessions) ConfirmTerminateOtherSessions() {}
