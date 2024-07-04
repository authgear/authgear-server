package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaTakeNewPassword struct {
	JSONPointer             jsonpointer.T
	FlowRootObject          config.AuthenticationFlowObject
	IsBotProtectionRequired bool
}

var _ authflow.InputSchema = &InputSchemaTakeNewPassword{}

func (i *InputSchemaTakeNewPassword) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaTakeNewPassword) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaTakeNewPassword) SchemaBuilder() validation.SchemaBuilder {
	inputTakeNewPasswordSchemaBuilder := validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("new_password")

	inputTakeNewPasswordSchemaBuilder.Properties().Property(
		"new_password",
		validation.SchemaBuilder{}.Type(validation.TypeString),
	)

	if i.IsBotProtectionRequired {
		inputTakeNewPasswordSchemaBuilder = AddBotProtectionToExistingSchemaBuilder(inputTakeNewPasswordSchemaBuilder)
	}
	return inputTakeNewPasswordSchemaBuilder
}

func (i *InputSchemaTakeNewPassword) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputTakeNewPassword
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputTakeNewPassword struct {
	NewPassword   string                      `json:"new_password,omitempty"`
	BotProtection *InputTakeBotProtectionBody `json:"bot_protection,omitempty"`
}

var _ authflow.Input = &InputTakeNewPassword{}
var _ inputTakeNewPassword = &InputTakeNewPassword{}
var _ inputTakeBotProtection = &InputTakeNewPassword{}

func (*InputTakeNewPassword) Input() {}

func (i *InputTakeNewPassword) GetNewPassword() string {
	return i.NewPassword
}

func (i *InputTakeNewPassword) GetBotProtectionProvider() *InputTakeBotProtectionBody {
	return i.BotProtection
}

func (i *InputTakeNewPassword) GetBotProtectionProviderType() config.BotProtectionProviderType {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Type
}

func (i *InputTakeNewPassword) GetBotProtectionProviderResponse() string {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Response
}
