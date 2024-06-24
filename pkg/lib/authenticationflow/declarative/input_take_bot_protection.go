package declarative

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

// This input should be dynamically added, but not used directly.
type InputTakeBotProtection struct {
	Type config.BotProtectionProviderType `json:"type,omitempty"`
	// Response is specific to cloudflare, recaptchav2
	Response string `json:"response,omitempty"`
}

func NewInputTakeBotProtectionSchemaBuilder() validation.SchemaBuilder {
	b := validation.SchemaBuilder{}.Type(validation.TypeObject)
	b.Required("type")
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

	return b
}

func AddBotProtectionToExistingSchemaBuilder(sb validation.SchemaBuilder) validation.SchemaBuilder {
	sb.AddRequired("bot_protection")
	sb.Properties().Property(("bot_protection"), NewInputTakeBotProtectionSchemaBuilder())
	return sb
}
