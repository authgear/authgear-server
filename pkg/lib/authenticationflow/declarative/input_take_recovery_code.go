package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
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

type InputSchemaTakeRecoveryCode struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
}

var _ authflow.InputSchema = &InputSchemaTakeRecoveryCode{}

func (i *InputSchemaTakeRecoveryCode) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaTakeRecoveryCode) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (*InputSchemaTakeRecoveryCode) SchemaBuilder() validation.SchemaBuilder {
	return InputTakeRecoveryCodeSchemaBuilder
}

func (i *InputSchemaTakeRecoveryCode) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputTakeRecoveryCode
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputTakeRecoveryCode struct {
	RecoveryCode       string `json:"recovery_code,omitempty"`
	RequestDeviceToken bool   `json:"request_device_token,omitempty"`
}

var _ authflow.Input = &InputTakeRecoveryCode{}
var _ inputTakeRecoveryCode = &InputTakeRecoveryCode{}
var _ inputDeviceTokenRequested = &InputTakeRecoveryCode{}

func (*InputTakeRecoveryCode) Input() {}

func (i *InputTakeRecoveryCode) GetRecoveryCode() string {
	return i.RecoveryCode
}

func (i *InputTakeRecoveryCode) GetDeviceTokenRequested() bool {
	return i.RequestDeviceToken
}
