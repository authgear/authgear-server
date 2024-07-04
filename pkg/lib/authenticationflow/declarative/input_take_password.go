package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaTakePassword struct {
	JSONPointer             jsonpointer.T
	FlowRootObject          config.AuthenticationFlowObject
	IsBotProtectionRequired bool
}

var _ authflow.InputSchema = &InputSchemaTakePassword{}

func (i *InputSchemaTakePassword) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaTakePassword) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaTakePassword) SchemaBuilder() validation.SchemaBuilder {
	inputTakePasswordSchemaBuilder := validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("password")

	inputTakePasswordSchemaBuilder.Properties().Property(
		"password",
		validation.SchemaBuilder{}.Type(validation.TypeString),
	)
	inputTakePasswordSchemaBuilder.Properties().Property(
		"request_device_token",
		validation.SchemaBuilder{}.Type(validation.TypeBoolean),
	)
	if i.IsBotProtectionRequired {
		inputTakePasswordSchemaBuilder = AddBotProtectionToExistingSchemaBuilder(inputTakePasswordSchemaBuilder)
	}
	return inputTakePasswordSchemaBuilder
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
	Password           string                      `json:"password,omitempty"`
	RequestDeviceToken bool                        `json:"request_device_token,omitempty"`
	BotProtection      *InputTakeBotProtectionBody `json:"bot_protection,omitempty"`
}

var _ authflow.Input = &InputTakePassword{}
var _ inputTakePassword = &InputTakePassword{}
var _ inputDeviceTokenRequested = &InputTakePassword{}
var _ inputTakeBotProtection = &InputTakePassword{}

func (*InputTakePassword) Input() {}

func (i *InputTakePassword) GetPassword() string {
	return i.Password
}

func (i *InputTakePassword) GetDeviceTokenRequested() bool {
	return i.RequestDeviceToken
}

func (i *InputTakePassword) GetBotProtectionProvider() *InputTakeBotProtectionBody {
	return i.BotProtection
}

func (i *InputTakePassword) GetBotProtectionProviderType() config.BotProtectionProviderType {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Type
}

func (i *InputTakePassword) GetBotProtectionProviderResponse() string {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Response
}
