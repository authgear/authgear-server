package workflowconfig

import (
	"encoding/json"

	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var InputTakePasswordSchemaBuilder validation.SchemaBuilder

func init() {
	InputTakePasswordSchemaBuilder = validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("password")

	InputTakePasswordSchemaBuilder.Properties().Property(
		"password",
		validation.SchemaBuilder{}.Type(validation.TypeString),
	)
	InputTakePasswordSchemaBuilder.Properties().Property(
		"request_device_token",
		validation.SchemaBuilder{}.Type(validation.TypeBoolean),
	)
}

type InputTakePassword struct {
	Password           string `json:"password,omitempty"`
	RequestDeviceToken bool   `json:"request_device_token,omitempty"`
}

var _ workflow.InputSchema = &InputTakePassword{}
var _ workflow.Input = &InputTakePassword{}
var _ inputTakePassword = &InputTakePassword{}
var _ inputDeviceTokenRequested = &InputTakePassword{}

func (*InputTakePassword) SchemaBuilder() validation.SchemaBuilder {
	return InputTakePasswordSchemaBuilder
}

func (i *InputTakePassword) MakeInput(rawMessage json.RawMessage) (workflow.Input, error) {
	var input InputTakePassword
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

func (*InputTakePassword) Input() {}

func (i *InputTakePassword) GetPassword() string {
	return i.Password
}

func (i *InputTakePassword) GetDeviceTokenRequested() bool {
	return i.RequestDeviceToken
}
