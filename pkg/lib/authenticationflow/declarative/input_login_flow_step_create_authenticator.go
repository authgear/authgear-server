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

type InputSchemaLoginFlowStepCreateAuthenticator struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
	OneOf          []*config.AuthenticationFlowLoginFlowOneOf
}

var _ authflow.InputSchema = &InputSchemaLoginFlowStepCreateAuthenticator{}

func (i *InputSchemaLoginFlowStepCreateAuthenticator) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaLoginFlowStepCreateAuthenticator) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaLoginFlowStepCreateAuthenticator) SchemaBuilder() validation.SchemaBuilder {
	oneOf := []validation.SchemaBuilder{}

	for _, branch := range i.OneOf {
		branch := branch
		b := validation.SchemaBuilder{}
		required := []string{"authentication"}
		b.Properties().Property("authentication", validation.SchemaBuilder{}.Const(branch.Authentication))

		switch branch.Authentication {
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
			if branch.TargetStep == "" {
				// Then target is required
				required = append(required, "target")
				b.Properties().Property("target", validation.SchemaBuilder{}.Type(validation.TypeString))
			}
			b.Required(required...)
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

func (i *InputSchemaLoginFlowStepCreateAuthenticator) MakeInput(ctx context.Context, rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputLoginFlowStepCreateAuthenticator
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(ctx, rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputLoginFlowStepCreateAuthenticator struct {
	Authentication model.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
	NewPassword    string                                 `json:"new_password,omitempty"`
	Target         string                                 `json:"target,omitempty"`
}

var _ authflow.Input = &InputLoginFlowStepCreateAuthenticator{}
var _ inputTakeAuthenticationMethod = &InputLoginFlowStepCreateAuthenticator{}
var _ inputTakeOOBOTPTarget = &InputLoginFlowStepCreateAuthenticator{}
var _ inputTakeNewPassword = &InputLoginFlowStepCreateAuthenticator{}

func (i *InputLoginFlowStepCreateAuthenticator) Input() {}

func (i *InputLoginFlowStepCreateAuthenticator) GetAuthenticationMethod() model.AuthenticationFlowAuthentication {
	return i.Authentication
}

func (i *InputLoginFlowStepCreateAuthenticator) GetTarget() string {
	return i.Target
}

func (i *InputLoginFlowStepCreateAuthenticator) GetNewPassword() string {
	return i.NewPassword
}
