package workflowconfig

import (
	"encoding/json"

	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
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

type InputTakeTOTP struct {
	Code               string `json:"code,omitempty"`
	RequestDeviceToken bool   `json:"request_device_token,omitempty"`
}

var _ workflow.InputSchema = &InputTakeTOTP{}
var _ workflow.Input = &InputTakeTOTP{}
var _ inputTakeTOTP = &InputTakeTOTP{}
var _ inputDeviceTokenRequested = &InputTakeTOTP{}

func (*InputTakeTOTP) SchemaBuilder() validation.SchemaBuilder {
	return InputTakeTOTPSchemaBuilder
}

func (i *InputTakeTOTP) MakeInput(rawMessage json.RawMessage) (workflow.Input, error) {
	var input InputTakeTOTP
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

func (*InputTakeTOTP) Input() {}

func (i *InputTakeTOTP) GetCode() string {
	return i.Code
}

func (i *InputTakeTOTP) GetDeviceTokenRequested() bool {
	return i.RequestDeviceToken
}
