package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var InputTakeTOTPSchemaBuilder validation.SchemaBuilder

func init() {
	InputTakeTOTPSchemaBuilder = validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("code")

	InputTakeTOTPSchemaBuilder.Properties().Property(
		"code",
		validation.SchemaBuilder{}.Type(validation.TypeString),
	)
	InputTakeTOTPSchemaBuilder.Properties().Property(
		"request_device_token",
		validation.SchemaBuilder{}.Type(validation.TypeBoolean),
	)
}

type InputSchemaTakeTOTP struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
}

var _ authflow.InputSchema = &InputSchemaTakeTOTP{}

func (i *InputSchemaTakeTOTP) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaTakeTOTP) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (*InputSchemaTakeTOTP) SchemaBuilder() validation.SchemaBuilder {
	return InputTakeTOTPSchemaBuilder
}

func (i *InputSchemaTakeTOTP) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputTakeTOTP
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputTakeTOTP struct {
	Code               string `json:"code,omitempty"`
	RequestDeviceToken bool   `json:"request_device_token,omitempty"`
}

var _ authflow.Input = &InputTakeTOTP{}
var _ inputTakeTOTP = &InputTakeTOTP{}
var _ inputDeviceTokenRequested = &InputTakeTOTP{}

func (*InputTakeTOTP) Input() {}

func (i *InputTakeTOTP) GetCode() string {
	return i.Code
}

func (i *InputTakeTOTP) GetDeviceTokenRequested() bool {
	return i.RequestDeviceToken
}
