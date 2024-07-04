package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var InputTakeBotProtectionSchemaBuilder validation.SchemaBuilder
var InputTakeBotProtectionBodySchemaBuilder validation.SchemaBuilder

func init() {
	InputTakeBotProtectionBodySchemaBuilder = validation.SchemaBuilder{}.Type(validation.TypeObject)
	InputTakeBotProtectionBodySchemaBuilder.Required("type")
	InputTakeBotProtectionBodySchemaBuilder.Properties().Property("type", validation.SchemaBuilder{}.Type(validation.TypeString).Enum(config.BotProtectionProviderTypeCloudflare, config.BotProtectionProviderTypeRecaptchaV2))
	InputTakeBotProtectionBodySchemaBuilder.Properties().Property("response", validation.SchemaBuilder{}.Type(validation.TypeString))

	// require "response" if type is in {"cloudflare", "recaptchav2"}
	allOf := validation.SchemaBuilder{}
	if_ := validation.SchemaBuilder{}
	if_.Properties().Property("type", validation.SchemaBuilder{}.Enum(config.BotProtectionProviderTypeCloudflare, config.BotProtectionProviderTypeRecaptchaV2))
	if_.Required("type")
	then_ := validation.SchemaBuilder{}
	then_.Required("response", "type")
	allOf.If(if_).Then(then_)

	InputTakeBotProtectionBodySchemaBuilder.AllOf(allOf)
	InputTakeBotProtectionSchemaBuilder = validation.SchemaBuilder{}.Type(validation.TypeObject)
	InputTakeBotProtectionSchemaBuilder = AddBotProtectionToExistingSchemaBuilder(InputTakeBotProtectionSchemaBuilder)
}

type InputSchemaTakeBotProtection struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
}

var _ authflow.InputSchema = &InputSchemaTakeBotProtection{}

func (i *InputSchemaTakeBotProtection) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaTakeBotProtection) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (*InputSchemaTakeBotProtection) SchemaBuilder() validation.SchemaBuilder {
	return InputTakeBotProtectionSchemaBuilder
}

func (i *InputSchemaTakeBotProtection) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputTakeBotProtection
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputTakeBotProtection struct {
	BotProtection *InputTakeBotProtectionBody `json:"bot_protection,omitempty"`
}

var _ authflow.Input = &InputTakeBotProtection{}
var _ inputTakeBotProtection = &InputTakeBotProtection{}

func (*InputTakeBotProtection) Input() {}

func (i *InputTakeBotProtection) GetBotProtectionProvider() *InputTakeBotProtectionBody {
	return i.BotProtection
}

func (i *InputTakeBotProtection) GetBotProtectionProviderType() config.BotProtectionProviderType {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Type
}

func (i *InputTakeBotProtection) GetBotProtectionProviderResponse() string {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Response
}

type InputTakeBotProtectionBody struct {
	Type config.BotProtectionProviderType `json:"type,omitempty"`
	// Response is specific to cloudflare, recaptchav2
	Response string `json:"response,omitempty"`
}

func AddBotProtectionToExistingSchemaBuilder(sb validation.SchemaBuilder) validation.SchemaBuilder {
	sb.AddRequired("bot_protection")
	sb.Properties().Property(("bot_protection"), InputTakeBotProtectionBodySchemaBuilder)
	return sb
}
