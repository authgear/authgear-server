package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var InputCreateDeviceTokenSchemaBuilder validation.SchemaBuilder

func init() {
	InputCreateDeviceTokenSchemaBuilder = validation.SchemaBuilder{}.
		Type(validation.TypeObject)

	InputCreateDeviceTokenSchemaBuilder.Properties().Property(
		"request_device_token",
		validation.SchemaBuilder{}.Type(validation.TypeBoolean),
	)
}

type InputSchemaCreateDeviceToken struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
}

var _ authflow.InputSchema = &InputSchemaCreateDeviceToken{}

func (i *InputSchemaCreateDeviceToken) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaCreateDeviceToken) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (*InputSchemaCreateDeviceToken) SchemaBuilder() validation.SchemaBuilder {
	return InputCreateDeviceTokenSchemaBuilder
}

func (i *InputSchemaCreateDeviceToken) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputCreateDeviceToken
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputCreateDeviceToken struct {
	RequestDeviceToken bool `json:"request_device_token,omitempty"`
}

var _ authflow.Input = &InputCreateDeviceToken{}
var _ inputDeviceTokenRequested = &InputCreateDeviceToken{}

func (*InputCreateDeviceToken) Input() {}

func (i *InputCreateDeviceToken) GetDeviceTokenRequested() bool {
	return i.RequestDeviceToken
}
