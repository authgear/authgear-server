package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaBotProtectionVerification struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
}

var _ authflow.InputSchema = &InputSchemaBotProtectionVerification{}

func (i *InputSchemaBotProtectionVerification) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaBotProtectionVerification) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaBotProtectionVerification) SchemaBuilder() validation.SchemaBuilder {
	b := validation.SchemaBuilder{}.Type(validation.TypeObject)
	b.Required("type")
	b.AdditionalPropertiesFalse()
	b.Properties().Property("type", validation.SchemaBuilder{}.Type(validation.TypeString).Enum(config.BotProtectionProviderTypeCloudflare, config.BotProtectionProviderTypeRecaptchaV2))
	b.Properties().Property("response", validation.SchemaBuilder{}.Type(validation.TypeString))

	// require "response" if type is in {"cloudflare", "recaptchav2"}
	allOf := validation.SchemaBuilder{}
	if_ := validation.SchemaBuilder{}
	if_.Properties().Property("type", validation.SchemaBuilder{}.Enum(config.BotProtectionProviderTypeCloudflare, config.BotProtectionProviderTypeRecaptchaV2))
	if_.Required("type")
	then_ := validation.SchemaBuilder{}
	then_.Required("response", "type")
	allOf.If(if_).Then(then_)
	b.AllOf(allOf)

	bRoot := validation.SchemaBuilder{}.Type(validation.TypeObject)
	bRoot.Required("bot_protection")
	bRoot.AdditionalPropertiesFalse()
	bRoot.Properties().Property("bot_protection", b)
	return bRoot
}

func (i *InputSchemaBotProtectionVerification) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputBotProtectionVerification
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputBotProtectionVerification struct {
	BotProtection *InputBotProtectionVerificationInfo `json:"bot_protection,omitempty"`
}

type InputBotProtectionVerificationInfo struct {
	Type config.BotProtectionProviderType `json:"type,omitempty"`

	// Response is specific to cloudflare, recaptchav2
	Response string `json:"response,omitempty"`
}

var _ authflow.Input = &InputBotProtectionVerification{}
var _ inputBotProtectionVerification = &InputBotProtectionVerification{}

func (*InputBotProtectionVerification) Input() {}

func (i *InputBotProtectionVerification) GetBotProtectionProviderType() config.BotProtectionProviderType {
	return i.BotProtection.Type
}

func (i *InputBotProtectionVerification) GetBotProtectionProviderResponse() string {
	return i.BotProtection.Response
}
