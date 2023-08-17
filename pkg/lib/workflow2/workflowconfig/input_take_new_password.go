package workflowconfig

import (
	"encoding/json"

	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var InputTakeNewPasswordSchemaBuilder validation.SchemaBuilder

func init() {
	InputTakeNewPasswordSchemaBuilder = validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("new_password")

	InputTakeNewPasswordSchemaBuilder.Properties().Property(
		"new_password",
		validation.SchemaBuilder{}.Type(validation.TypeString),
	)
}

type InputTakeNewPassword struct {
	NewPassword string `json:"new_password,omitempty"`
}

var _ workflow.InputSchema = &InputTakeNewPassword{}
var _ workflow.Input = &InputTakeNewPassword{}
var _ inputTakeNewPassword = &InputTakeNewPassword{}

func (*InputTakeNewPassword) SchemaBuilder() validation.SchemaBuilder {
	return InputTakeNewPasswordSchemaBuilder
}

func (i *InputTakeNewPassword) MakeInput(rawMessage json.RawMessage) (workflow.Input, error) {
	var input InputTakeNewPassword
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

func (*InputTakeNewPassword) Input() {}

func (i *InputTakeNewPassword) GetNewPassword() string {
	return i.NewPassword
}
