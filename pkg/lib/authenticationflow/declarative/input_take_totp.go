package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaTakeTOTP struct {
	JSONPointer             jsonpointer.T
	FlowRootObject          config.AuthenticationFlowObject
	IsBotProtectionRequired bool
}

var _ authflow.InputSchema = &InputSchemaTakeTOTP{}

func (i *InputSchemaTakeTOTP) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaTakeTOTP) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaTakeTOTP) SchemaBuilder() validation.SchemaBuilder {
	inputTakeTOTPSchemaBuilder := validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("code")

	inputTakeTOTPSchemaBuilder.Properties().Property(
		"code",
		validation.SchemaBuilder{}.Type(validation.TypeString),
	)
	inputTakeTOTPSchemaBuilder.Properties().Property(
		"request_device_token",
		validation.SchemaBuilder{}.Type(validation.TypeBoolean),
	)

	if i.IsBotProtectionRequired {
		inputTakeTOTPSchemaBuilder = AddBotProtectionToExistingSchemaBuilder(inputTakeTOTPSchemaBuilder)
	}
	return inputTakeTOTPSchemaBuilder
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
	Code               string                  `json:"code,omitempty"`
	RequestDeviceToken bool                    `json:"request_device_token,omitempty"`
	BotProtection      *InputTakeBotProtection `json:"bot_protection,omitempty"`
}

var _ authflow.Input = &InputTakeTOTP{}
var _ inputTakeTOTP = &InputTakeTOTP{}
var _ inputDeviceTokenRequested = &InputTakeTOTP{}
var _ inputTakeBotProtection = &InputTakeTOTP{}

func (*InputTakeTOTP) Input() {}

func (i *InputTakeTOTP) GetCode() string {
	return i.Code
}

func (i *InputTakeTOTP) GetDeviceTokenRequested() bool {
	return i.RequestDeviceToken
}

func (i *InputTakeTOTP) GetBotProtectionProvider() *InputTakeBotProtection {
	return i.BotProtection
}

func (i *InputTakeTOTP) GetBotProtectionProviderType() config.BotProtectionProviderType {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Type
}

func (i *InputTakeTOTP) GetBotProtectionProviderResponse() string {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Response
}
