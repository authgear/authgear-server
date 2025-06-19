package declarative

import (
	"context"
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/slice"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaUseAuthenticatorOOBOTP struct {
	JSONPointer               jsonpointer.T
	FlowRootObject            config.AuthenticationFlowObject
	Options                   []AuthenticateOption
	ShouldBypassBotProtection bool
	BotProtectionCfg          *config.BotProtectionConfig
}

var _ authflow.InputSchema = &InputSchemaUseAuthenticatorOOBOTP{}

func (i *InputSchemaUseAuthenticatorOOBOTP) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaUseAuthenticatorOOBOTP) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaUseAuthenticatorOOBOTP) SchemaBuilder() validation.SchemaBuilder {
	optionSchemaPairs := []struct {
		index  int
		schema validation.SchemaBuilder
	}{}

	for index, option := range i.Options {
		index := index
		option := option

		addPair := func() {
			optionSchema := validation.SchemaBuilder{}
			if !i.ShouldBypassBotProtection && i.BotProtectionCfg != nil && option.isBotProtectionRequired() {
				optionSchema = AddBotProtectionToExistingSchemaBuilder(optionSchema, i.BotProtectionCfg)
			}
			optionSchemaPairs = append(optionSchemaPairs, struct {
				index  int
				schema validation.SchemaBuilder
			}{
				index:  index,
				schema: optionSchema,
			})
		}

		switch option.Authentication {
		case model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case model.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			fallthrough
		case model.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			addPair()
		default:
			break
		}
	}

	b := validation.SchemaBuilder{}.
		Type(validation.TypeObject)
	indice := []int{}
	allOfs := []validation.SchemaBuilder{}
	for _, pair := range optionSchemaPairs {
		indice = append(indice, pair.index)
		if_ := validation.SchemaBuilder{}
		if_.Properties().Property("index", validation.SchemaBuilder{}.Const(pair.index))
		ifSchema := validation.SchemaBuilder{}
		ifSchema.If(if_).Then(pair.schema)
		allOfs = append(allOfs, ifSchema)
	}

	b.Properties().Property("index", validation.SchemaBuilder{}.
		Type(validation.TypeInteger).
		Enum(slice.Cast[int, interface{}](indice)...),
	)
	b.AllOf(allOfs...)

	return b
}

func (i *InputSchemaUseAuthenticatorOOBOTP) MakeInput(ctx context.Context, rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputUseAuthenticatorOOBOTP
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(ctx, rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputUseAuthenticatorOOBOTP struct {
	Index         int                         `json:"index"`
	BotProtection *InputTakeBotProtectionBody `json:"bot_protection,omitempty"`
}

var _ authflow.Input = &InputUseAuthenticatorOOBOTP{}
var _ inputTakeAuthenticationOptionIndex = &InputUseAuthenticatorOOBOTP{}
var _ inputTakeBotProtection = &InputUseAuthenticatorOOBOTP{}

func (*InputUseAuthenticatorOOBOTP) Input() {}

func (i *InputUseAuthenticatorOOBOTP) GetIndex() int {
	return i.Index
}

func (i *InputUseAuthenticatorOOBOTP) GetBotProtectionProvider() *InputTakeBotProtectionBody {
	return i.BotProtection
}

func (i *InputUseAuthenticatorOOBOTP) GetBotProtectionProviderType() config.BotProtectionProviderType {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Type
}

func (i *InputUseAuthenticatorOOBOTP) GetBotProtectionProviderResponse() string {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Response
}
