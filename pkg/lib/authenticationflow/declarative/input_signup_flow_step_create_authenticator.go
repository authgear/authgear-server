package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaSignupFlowStepCreateAuthenticator struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
	OneOf          []*config.AuthenticationFlowSignupFlowOneOf
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
		case config.AuthenticationFlowAuthenticationPrimaryPasskey:
			// Cannot create passkey in this step.
			break
		default:
			break
		}
	}

	return validation.SchemaBuilder{}.Type(validation.TypeObject).OneOf(oneOf...)
}

func (i *InputSchemaSignupFlowStepCreateAuthenticator) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputSignupFlowStepCreateAuthenticator
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputSignupFlowStepCreateAuthenticator struct {
	Authentication config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
	NewPassword    string                                  `json:"new_password,omitempty"`
	Target         string                                  `json:"target,omitempty"`
}

var _ authflow.Input = &InputSignupFlowStepCreateAuthenticator{}
var _ inputTakeAuthenticationMethod = &InputSignupFlowStepCreateAuthenticator{}
var _ inputTakeOOBOTPTarget = &InputSignupFlowStepCreateAuthenticator{}
var _ inputTakeNewPassword = &InputSignupFlowStepCreateAuthenticator{}

func (i *InputSignupFlowStepCreateAuthenticator) Input() {}

func (i *InputSignupFlowStepCreateAuthenticator) GetAuthenticationMethod() config.AuthenticationFlowAuthentication {
	return i.Authentication
}

func (i *InputSignupFlowStepCreateAuthenticator) GetTarget() string {
	return i.Target
}

func (i *InputSignupFlowStepCreateAuthenticator) GetNewPassword() string {
	return i.NewPassword
}
