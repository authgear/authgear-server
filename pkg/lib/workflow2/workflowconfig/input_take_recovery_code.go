package workflowconfig

import (
	"encoding/json"

	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var InputTakeRecoveryCodeSchemaBuilder validation.SchemaBuilder

func init() {
	InputTakeRecoveryCodeSchemaBuilder = validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("recovery_code")

	InputTakeRecoveryCodeSchemaBuilder.Properties().Property(
		"recovery_code",
		validation.SchemaBuilder{}.Type(validation.TypeString),
	)
	InputTakeRecoveryCodeSchemaBuilder.Properties().Property(
		"request_device_token",
		validation.SchemaBuilder{}.Type(validation.TypeBoolean),
	)
}

type InputTakeRecoveryCode struct {
	RecoveryCode       string `json:"recovery_code,omitempty"`
	RequestDeviceToken bool   `json:"request_device_token,omitempty"`
}

var _ workflow.InputSchema = &InputTakeRecoveryCode{}
var _ workflow.Input = &InputTakeRecoveryCode{}
var _ inputTakeRecoveryCode = &InputTakeRecoveryCode{}
var _ inputDeviceTokenRequested = &InputTakeRecoveryCode{}

func (*InputTakeRecoveryCode) SchemaBuilder() validation.SchemaBuilder {
	return InputTakeRecoveryCodeSchemaBuilder
}

func (i *InputTakeRecoveryCode) MakeInput(rawMessage json.RawMessage) (workflow.Input, error) {
	var input InputTakeRecoveryCode
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

func (*InputTakeRecoveryCode) Input() {}

func (i *InputTakeRecoveryCode) GetRecoveryCode() string {
	return i.RecoveryCode
}

func (i *InputTakeRecoveryCode) GetDeviceTokenRequested() bool {
	return i.RequestDeviceToken
}
