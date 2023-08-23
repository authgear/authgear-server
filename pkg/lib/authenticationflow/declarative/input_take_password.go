package declarative

import (
	"encoding/json"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
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

var _ authflow.InputSchema = &InputTakePassword{}
var _ authflow.Input = &InputTakePassword{}
var _ inputTakePassword = &InputTakePassword{}
var _ inputDeviceTokenRequested = &InputTakePassword{}

func (*InputTakePassword) SchemaBuilder() validation.SchemaBuilder {
	return InputTakePasswordSchemaBuilder
}

func (i *InputTakePassword) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
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
