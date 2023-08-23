package declarative

import (
	"encoding/json"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaSignupFlowStepAuthenticate struct {
	OneOf []*config.AuthenticationFlowSignupFlowOneOf
}

var _ authflow.InputSchema = &InputSchemaSignupFlowStepAuthenticate{}

func (i *InputSchemaSignupFlowStepAuthenticate) SchemaBuilder() validation.SchemaBuilder {
	oneOf := []validation.SchemaBuilder{}

	for _, branch := range i.OneOf {
		branch := branch
		b := validation.SchemaBuilder{}
		required := []string{"authentication"}
		b.Properties().Property("authentication", validation.SchemaBuilder{}.Const(branch.Authentication))

		switch branch.Authentication {
		case config.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryPassword:
			// Require new_password
			required = append(required, "new_password")
			b.Properties().Property("new_password", validation.SchemaBuilder{}.Type(validation.TypeString))
			b.Required(required...)
			oneOf = append(oneOf, b)
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			// No other property is required.
			b.Required(required...)
			oneOf = append(oneOf, b)
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			if branch.TargetStep == "" {
				// Then target is required
				required = append(required, "target")
				b.Properties().Property("target", validation.SchemaBuilder{}.Type(validation.TypeString))
			}
			b.Required(required...)
			oneOf = append(oneOf, b)
		default:
			break
		}
	}

	return validation.SchemaBuilder{}.Type(validation.TypeObject).OneOf(oneOf...)
}

func (i *InputSchemaSignupFlowStepAuthenticate) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputSignupFlowStepAuthenticate
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputSignupFlowStepAuthenticate struct {
	Authentication config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
	NewPassword    string                                  `json:"new_password,omitempty"`
	Target         string                                  `json:"target,omitempty"`
}

var _ authflow.Input = &InputSignupFlowStepAuthenticate{}
var _ inputTakeAuthenticationMethod = &InputSignupFlowStepAuthenticate{}
var _ inputTakeOOBOTPTarget = &InputSignupFlowStepAuthenticate{}
var _ inputTakeNewPassword = &InputSignupFlowStepAuthenticate{}

func (i *InputSignupFlowStepAuthenticate) Input() {}

func (i *InputSignupFlowStepAuthenticate) GetAuthenticationMethod() config.AuthenticationFlowAuthentication {
	return i.Authentication
}

func (i *InputSignupFlowStepAuthenticate) GetTarget() string {
	return i.Target
}

func (i *InputSignupFlowStepAuthenticate) GetNewPassword() string {
	return i.NewPassword
}
