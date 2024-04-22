package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
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

type InputSchemaTakePassword struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
}

var _ authflow.InputSchema = &InputSchemaTakePassword{}

func (i *InputSchemaTakePassword) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaTakePassword) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (*InputSchemaTakePassword) SchemaBuilder() validation.SchemaBuilder {
	return InputTakePasswordSchemaBuilder
}

func (i *InputSchemaTakePassword) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputTakePassword
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputTakePassword struct {
	Password           string `json:"password,omitempty"`
	RequestDeviceToken bool   `json:"request_device_token,omitempty"`
}

var _ authflow.Input = &InputTakePassword{}
var _ inputTakePassword = &InputTakePassword{}
var _ inputDeviceTokenRequested = &InputTakePassword{}

func (*InputTakePassword) Input() {}

func (i *InputTakePassword) GetPassword() string {
	return i.Password
}

func (i *InputTakePassword) GetDeviceTokenRequested() bool {
	return i.RequestDeviceToken
}
