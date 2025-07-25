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

type InputSchemaReauthFlowStepAuthenticate struct {
	JSONPointer               jsonpointer.T
	FlowRootObject            config.AuthenticationFlowObject
	Options                   []AuthenticateOption
	ShouldBypassBotProtection bool
	BotProtectionCfg          *config.BotProtectionConfig
}

var _ authflow.InputSchema = &InputSchemaReauthFlowStepAuthenticate{}

func (i *InputSchemaReauthFlowStepAuthenticate) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaReauthFlowStepAuthenticate) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaReauthFlowStepAuthenticate) SchemaBuilder() validation.SchemaBuilder {
	oneOf := []validation.SchemaBuilder{}

	for index, option := range i.Options {
		index := index
		option := option

		b := validation.SchemaBuilder{}
		required := []string{"authentication"}
		b.Properties().Property("authentication", validation.SchemaBuilder{}.Const(option.Authentication))

		requireString := func(key string) {
			required = append(required, key)
			b.Properties().Property(key, validation.SchemaBuilder{}.Type(validation.TypeString))
		}
		requireIndex := func() {
			required = append(required, "index")
			b.Properties().Property("index", validation.SchemaBuilder{}.
				Type(validation.TypeInteger).
				Const(index),
			)
		}
		requireBotProtection := func() {
			required = append(required, "bot_protection")
			b.Properties().Property("bot_protection", NewBotProtectionBodySchemaBuilder(i.BotProtectionCfg))
		}
		mayRequireChannel := func() {
			if len(option.Channels) > 1 {
				required = append(required, "channel")
				b.Properties().Property("channel", validation.SchemaBuilder{}.
					Type(validation.TypeString).
					Enum(slice.Cast[model.AuthenticatorOOBChannel, interface{}](option.Channels)...),
				)
			}
		}

		setRequiredAndAppendOneOf := func() {
			b.Required(required...)
			oneOf = append(oneOf, b)
		}

		if !i.ShouldBypassBotProtection && i.BotProtectionCfg != nil && option.isBotProtectionRequired() {
			requireBotProtection()
		}

		switch option.Authentication {
		case model.AuthenticationFlowAuthenticationPrimaryPassword:
			requireString("password")
			setRequiredAndAppendOneOf()
		case model.AuthenticationFlowAuthenticationPrimaryPasskey:
			required = append(required, "assertion_response")
			b.Properties().Property("assertion_response", passkeyAssertionResponseSchemaBuilder)
			setRequiredAndAppendOneOf()

		case model.AuthenticationFlowAuthenticationSecondaryPassword:
			requireString("password")
			setRequiredAndAppendOneOf()
		case model.AuthenticationFlowAuthenticationSecondaryTOTP:
			requireString("code")
			setRequiredAndAppendOneOf()
		case model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			requireIndex()
			mayRequireChannel()
			setRequiredAndAppendOneOf()
		case model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			requireIndex()
			mayRequireChannel()
			setRequiredAndAppendOneOf()
		case model.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			requireIndex()
			mayRequireChannel()
			setRequiredAndAppendOneOf()
		case model.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			requireIndex()
			mayRequireChannel()
			setRequiredAndAppendOneOf()
		default:
			break
		}
	}

	b := validation.SchemaBuilder{}.
		Type(validation.TypeObject)

	if len(oneOf) > 0 {
		b.OneOf(oneOf...)
	}

	return b
}

func (i *InputSchemaReauthFlowStepAuthenticate) MakeInput(ctx context.Context, rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputReauthFlowStepAuthenticate
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(ctx, rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputReauthFlowStepAuthenticate struct {
	Authentication model.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
	Password       string                                 `json:"password,omitempty"`
	Code           string                                 `json:"code,omitempty"`
	Index          int                                    `json:"index,omitempty"`
	Channel        model.AuthenticatorOOBChannel          `json:"channel,omitempty"`
}

var _ authflow.Input = &InputReauthFlowStepAuthenticate{}
var _ inputTakeAuthenticationMethod = &InputReauthFlowStepAuthenticate{}
var _ inputTakePassword = &InputReauthFlowStepAuthenticate{}
var _ inputTakeTOTP = &InputReauthFlowStepAuthenticate{}
var _ inputTakeAuthenticationOptionIndex = &InputReauthFlowStepAuthenticate{}
var _ inputTakeOOBOTPChannel = &InputReauthFlowStepAuthenticate{}

func (*InputReauthFlowStepAuthenticate) Input() {}

func (i *InputReauthFlowStepAuthenticate) GetAuthenticationMethod() model.AuthenticationFlowAuthentication {
	return i.Authentication
}

func (i *InputReauthFlowStepAuthenticate) GetPassword() string {
	return i.Password
}

func (i *InputReauthFlowStepAuthenticate) GetCode() string {
	return i.Code
}

func (i *InputReauthFlowStepAuthenticate) GetIndex() int {
	return i.Index
}

func (i *InputReauthFlowStepAuthenticate) GetChannel() model.AuthenticatorOOBChannel {
	return i.Channel
}
