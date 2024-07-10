package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaTakeRecoveryCode struct {
	JSONPointer             jsonpointer.T
	FlowRootObject          config.AuthenticationFlowObject
	IsBotProtectionRequired bool
	BotProtectionCfg        *config.BotProtectionConfig
}

var _ authflow.InputSchema = &InputSchemaTakeRecoveryCode{}

func (i *InputSchemaTakeRecoveryCode) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaTakeRecoveryCode) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaTakeRecoveryCode) SchemaBuilder() validation.SchemaBuilder {
	inputTakeRecoveryCodeSchemaBuilder := validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("recovery_code")

	inputTakeRecoveryCodeSchemaBuilder.Properties().Property(
		"recovery_code",
		validation.SchemaBuilder{}.Type(validation.TypeString),
	)
	inputTakeRecoveryCodeSchemaBuilder.Properties().Property(
		"request_device_token",
		validation.SchemaBuilder{}.Type(validation.TypeBoolean),
	)
	if i.IsBotProtectionRequired && i.BotProtectionCfg != nil {
		inputTakeRecoveryCodeSchemaBuilder = AddBotProtectionToExistingSchemaBuilder(inputTakeRecoveryCodeSchemaBuilder, i.BotProtectionCfg)
	}
	return inputTakeRecoveryCodeSchemaBuilder
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
	RecoveryCode       string                      `json:"recovery_code,omitempty"`
	RequestDeviceToken bool                        `json:"request_device_token,omitempty"`
	BotProtection      *InputTakeBotProtectionBody `json:"bot_protection,omitempty"`
}

var _ authflow.Input = &InputTakeRecoveryCode{}
var _ inputTakeRecoveryCode = &InputTakeRecoveryCode{}
var _ inputDeviceTokenRequested = &InputTakeRecoveryCode{}
var _ inputTakeBotProtection = &InputTakeRecoveryCode{}

func (*InputTakeRecoveryCode) Input() {}

func (i *InputTakeRecoveryCode) GetRecoveryCode() string {
	return i.RecoveryCode
}

func (i *InputTakeRecoveryCode) GetDeviceTokenRequested() bool {
	return i.RequestDeviceToken
}

func (i *InputTakeRecoveryCode) GetBotProtectionProvider() *InputTakeBotProtectionBody {
	return i.BotProtection
}

func (i *InputTakeRecoveryCode) GetBotProtectionProviderType() config.BotProtectionProviderType {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Type
}

func (i *InputTakeRecoveryCode) GetBotProtectionProviderResponse() string {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Response
}
