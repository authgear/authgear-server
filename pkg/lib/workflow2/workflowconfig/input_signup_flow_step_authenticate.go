package workflowconfig

import (
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/lib/config"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaSignupFlowStepAuthenticate struct {
	OneOf []*config.WorkflowSignupFlowOneOf
}

var _ workflow.InputSchema = &InputSchemaSignupFlowStepAuthenticate{}

func (i *InputSchemaSignupFlowStepAuthenticate) SchemaBuilder() validation.SchemaBuilder {
	oneOf := []validation.SchemaBuilder{}

	for _, branch := range i.OneOf {
		branch := branch
		b := validation.SchemaBuilder{}
		required := []string{"authentication"}
		b.Properties().Property("authentication", validation.SchemaBuilder{}.Const(branch.Authentication))

		switch branch.Authentication {
		case config.WorkflowAuthenticationMethodPrimaryPassword:
			fallthrough
		case config.WorkflowAuthenticationMethodSecondaryPassword:
			// Require new_password
			required = append(required, "new_password")
			b.Properties().Property("new_password", validation.SchemaBuilder{}.Type(validation.TypeString))
			b.Required(required...)
			oneOf = append(oneOf, b)
		case config.WorkflowAuthenticationMethodSecondaryTOTP:
			// No other property is required.
			b.Required(required...)
			oneOf = append(oneOf, b)
		case config.WorkflowAuthenticationMethodPrimaryOOBOTPEmail:
			fallthrough
		case config.WorkflowAuthenticationMethodPrimaryOOBOTPSMS:
			fallthrough
		case config.WorkflowAuthenticationMethodSecondaryOOBOTPEmail:
			fallthrough
		case config.WorkflowAuthenticationMethodSecondaryOOBOTPSMS:
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

func (i *InputSchemaSignupFlowStepAuthenticate) MakeInput(rawMessage json.RawMessage) (workflow.Input, error) {
	var input InputSignupFlowStepAuthenticate
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputSignupFlowStepAuthenticate struct {
	Authentication config.WorkflowAuthenticationMethod `json:"authentication,omitempty"`
	NewPassword    string                              `json:"new_password,omitempty"`
	Target         string                              `json:"target,omitempty"`
}

var _ workflow.Input = &InputSignupFlowStepAuthenticate{}
var _ inputTakeAuthenticationMethod = &InputSignupFlowStepAuthenticate{}
var _ inputTakeOOBOTPTarget = &InputSignupFlowStepAuthenticate{}
var _ inputTakeNewPassword = &InputSignupFlowStepAuthenticate{}

func (i *InputSignupFlowStepAuthenticate) Input() {}

func (i *InputSignupFlowStepAuthenticate) GetAuthenticationMethod() config.WorkflowAuthenticationMethod {
	return i.Authentication
}

func (i *InputSignupFlowStepAuthenticate) GetTarget() string {
	return i.Target
}

func (i *InputSignupFlowStepAuthenticate) GetNewPassword() string {
	return i.NewPassword
}
