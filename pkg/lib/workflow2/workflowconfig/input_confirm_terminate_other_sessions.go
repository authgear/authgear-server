package workflowconfig

import (
	"encoding/json"

	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
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

type InputConfirmTerminateOtherSessions struct{}

var _ workflow.InputSchema = &InputConfirmTerminateOtherSessions{}
var _ workflow.Input = &InputConfirmTerminateOtherSessions{}
var _ inputConfirmTerminateOtherSessions = &InputConfirmTerminateOtherSessions{}

func (*InputConfirmTerminateOtherSessions) SchemaBuilder() validation.SchemaBuilder {
	return InputConfirmTerminateOtherSessionsSchemaBuilder
}

func (i *InputConfirmTerminateOtherSessions) MakeInput(rawMessage json.RawMessage) (workflow.Input, error) {
	var input InputConfirmTerminateOtherSessions
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

func (*InputConfirmTerminateOtherSessions) Input() {}

func (*InputConfirmTerminateOtherSessions) ConfirmTerminateOtherSessions() {}
