package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaTakeBotProtection struct {
	JSONPointer      jsonpointer.T
	FlowRootObject   config.AuthenticationFlowObject
	BotProtectionCfg *config.BotProtectionConfig
}

var _ authflow.InputSchema = &InputSchemaTakeBotProtection{}

func (i *InputSchemaTakeBotProtection) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaTakeBotProtection) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaTakeBotProtection) SchemaBuilder() validation.SchemaBuilder {
	b := validation.SchemaBuilder{}.Type(validation.TypeObject)
	b = AddBotProtectionToExistingSchemaBuilder(b, i.BotProtectionCfg)
	return b
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

func NewBotProtectionBodySchemaBuilder(bpCfg *config.BotProtectionConfig) validation.SchemaBuilder {
	if bpCfg == nil || bpCfg.Provider == nil || bpCfg.Provider.Type == "" {
		panic("invalid bot protection config")
	}

	targetProviderType := bpCfg.Provider.Type
	b := validation.SchemaBuilder{}.Type(validation.TypeObject)
	b.Required("type")
	b.Properties().Property("type", validation.SchemaBuilder{}.Const(targetProviderType))

	switch targetProviderType {
	case config.BotProtectionProviderTypeCloudflare:
		b.Properties().Property("response", validation.SchemaBuilder{}.Type(validation.TypeString))
		b.AddRequired("response")
	case config.BotProtectionProviderTypeRecaptchaV2:
		b.Properties().Property("response", validation.SchemaBuilder{}.Type(validation.TypeString))
		b.AddRequired("response")
	}

	return b
}

func AddBotProtectionToExistingSchemaBuilder(sb validation.SchemaBuilder, bpCfg *config.BotProtectionConfig) validation.SchemaBuilder {
	sb.AddRequired("bot_protection")
	sb.Properties().Property(("bot_protection"), NewBotProtectionBodySchemaBuilder(bpCfg))
	return sb
}
