package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaStepAccountRecoveryVerifyCode struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
}

var _ authflow.InputSchema = &InputSchemaStepAccountRecoveryVerifyCode{}

func (i *InputSchemaStepAccountRecoveryVerifyCode) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaStepAccountRecoveryVerifyCode) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaStepAccountRecoveryVerifyCode) SchemaBuilder() validation.SchemaBuilder {
	resend := validation.SchemaBuilder{}.
		Required("resend")
	resend.Properties().Property(
		"resend", validation.SchemaBuilder{}.Type(validation.TypeBoolean).Const(true),
	)

	oneOf := []validation.SchemaBuilder{resend}

	code := validation.SchemaBuilder{}.
		Required("account_recovery_code")
	code.Properties().Property(
		"account_recovery_code", validation.SchemaBuilder{}.Type(validation.TypeString),
	)

	oneOf = append(oneOf, code)

	return validation.SchemaBuilder{}.Type(validation.TypeObject).OneOf(oneOf...)
}

func (i *InputSchemaStepAccountRecoveryVerifyCode) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputStepAccountRecoveryVerifyCode
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputStepAccountRecoveryVerifyCode struct {
	AccountRecoveryCode string `json:"account_recovery_code,omitempty"`
	Resend              bool   `json:"resend,omitempty"`
	Check               bool   `json:"check,omitempty"`
	RequestDeviceToken  bool   `json:"request_device_token,omitempty"`
}

var _ authflow.Input = &InputStepAccountRecoveryVerifyCode{}
var _ inputStepAccountRecoveryVerifyCode = &InputStepAccountRecoveryVerifyCode{}

func (*InputStepAccountRecoveryVerifyCode) Input() {}

func (i *InputStepAccountRecoveryVerifyCode) IsCode() bool {
	return i.AccountRecoveryCode != ""
}

func (i *InputStepAccountRecoveryVerifyCode) GetCode() string {
	return i.AccountRecoveryCode
}

func (i *InputStepAccountRecoveryVerifyCode) IsResend() bool {
	return i.Resend
}
