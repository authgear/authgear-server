package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

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
	indice := []int{}
	b := validation.SchemaBuilder{}.
		Type(validation.TypeObject)
	for index, option := range i.Options {
		index := index
		option := option

		switch option.Authentication {
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			indice = append(indice, index)
		default:
			break
		}
		if !i.ShouldBypassBotProtection && i.BotProtectionCfg != nil && option.isBotProtectionRequired() {
			b = AddBotProtectionToExistingSchemaBuilder(b, i.BotProtectionCfg)
		}
	}

	b.Properties().Property("index", validation.SchemaBuilder{}.
		Type(validation.TypeInteger).
		Enum(slice.Cast[int, interface{}](indice)...),
	)

	return b
}

func (i *InputSchemaUseAuthenticatorOOBOTP) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputUseAuthenticatorOOBOTP
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
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
