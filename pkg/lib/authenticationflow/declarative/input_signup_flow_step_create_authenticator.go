package declarative

import (
	"context"
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaSignupFlowStepCreateAuthenticator struct {
	JSONPointer               jsonpointer.T
	FlowRootObject            config.AuthenticationFlowObject
	Options                   []CreateAuthenticatorOption
	ShouldBypassBotProtection bool
	BotProtectionCfg          *config.BotProtectionConfig
}

var _ authflow.InputSchema = &InputSchemaSignupFlowStepCreateAuthenticator{}

func (i *InputSchemaSignupFlowStepCreateAuthenticator) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaSignupFlowStepCreateAuthenticator) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaSignupFlowStepCreateAuthenticator) SchemaBuilder() validation.SchemaBuilder {
	oneOf := []validation.SchemaBuilder{}

	for _, option := range i.Options {
		option := option
		b := validation.SchemaBuilder{}
		required := []string{"authentication"}
		b.Properties().Property("authentication", validation.SchemaBuilder{}.Const(option.Authentication))
		switch option.Authentication {
		case model.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case model.AuthenticationFlowAuthenticationSecondaryPassword:
			// Require new_password
			required = append(required, "new_password")
			b.Properties().Property("new_password", validation.SchemaBuilder{}.Type(validation.TypeString))
			b.Required(required...)
			oneOf = append(oneOf, b)
		case model.AuthenticationFlowAuthenticationSecondaryTOTP:
			// No other property is required.
			b.Required(required...)
			oneOf = append(oneOf, b)
		case model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case model.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			fallthrough
		case model.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			if option.Target == nil {
				// Then target is required
				required = append(required, "target")
				b.Properties().Property("target", validation.SchemaBuilder{}.Type(validation.TypeString))
			}
			b.Required(required...)
			if !i.ShouldBypassBotProtection && i.BotProtectionCfg != nil && option.isBotProtectionRequired() {
				b = AddBotProtectionToExistingSchemaBuilder(b, i.BotProtectionCfg)
			}
			oneOf = append(oneOf, b)
		case model.AuthenticationFlowAuthenticationPrimaryPasskey:
			// Cannot create passkey in this step.
			break
		default:
			break
		}
	}

	return validation.SchemaBuilder{}.Type(validation.TypeObject).OneOf(oneOf...)
}

func (i *InputSchemaSignupFlowStepCreateAuthenticator) MakeInput(ctx context.Context, rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputSignupFlowStepCreateAuthenticator
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(ctx, rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputSignupFlowStepCreateAuthenticator struct {
	Authentication model.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
	NewPassword    string                                 `json:"new_password,omitempty"`
	Target         string                                 `json:"target,omitempty"`

	BotProtection *InputTakeBotProtectionBody `json:"bot_protection,omitempty"`
}

var _ authflow.Input = &InputSignupFlowStepCreateAuthenticator{}
var _ inputTakeAuthenticationMethod = &InputSignupFlowStepCreateAuthenticator{}
var _ inputTakeOOBOTPTarget = &InputSignupFlowStepCreateAuthenticator{}
var _ inputTakeNewPassword = &InputSignupFlowStepCreateAuthenticator{}
var _ inputTakeBotProtection = &InputSignupFlowStepCreateAuthenticator{}

func (i *InputSignupFlowStepCreateAuthenticator) Input() {}

func (i *InputSignupFlowStepCreateAuthenticator) GetAuthenticationMethod() model.AuthenticationFlowAuthentication {
	return i.Authentication
}

func (i *InputSignupFlowStepCreateAuthenticator) GetTarget() string {
	return i.Target
}

func (i *InputSignupFlowStepCreateAuthenticator) GetNewPassword() string {
	return i.NewPassword
}

func (i *InputSignupFlowStepCreateAuthenticator) GetBotProtectionProvider() *InputTakeBotProtectionBody {
	return i.BotProtection
}

func (i *InputSignupFlowStepCreateAuthenticator) GetBotProtectionProviderType() config.BotProtectionProviderType {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Type
}

func (i *InputSignupFlowStepCreateAuthenticator) GetBotProtectionProviderResponse() string {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Response
}
