package workflowconfig

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

type IntentSignupFlowStepAuthenticateTarget interface {
	GetOOBOTPClaims(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (map[model.ClaimName]string, error)
}

func init() {
	workflow.RegisterIntent(&IntentSignupFlowStepAuthenticate{})
}

type IntentSignupFlowStepAuthenticate struct {
	SignupFlow  string        `json:"signup_flow,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	StepID      string        `json:"step_id,omitempty"`
	UserID      string        `json:"user_id,omitempty"`
}

var _ WorkflowStep = &IntentSignupFlowStepAuthenticate{}

func (i *IntentSignupFlowStepAuthenticate) GetID() string {
	return i.StepID
}

func (i *IntentSignupFlowStepAuthenticate) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ IntentSignupFlowStepVerifyTarget = &IntentSignupFlowStepAuthenticate{}

func (*IntentSignupFlowStepAuthenticate) GetVerifiableClaims(_ context.Context, _ *workflow.Dependencies, workflows workflow.Workflows) (map[model.ClaimName]string, error) {
	m, ok := FindMilestone[MilestoneDoCreateAuthenticator](workflows.Nearest)
	if !ok {
		return nil, fmt.Errorf("MilestoneDoCreateAuthenticator cannot be found in IntentSignupFlowStepAuthenticate")
	}

	info := m.MilestoneDoCreateAuthenticator()

	return info.StandardClaims(), nil
}

func (*IntentSignupFlowStepAuthenticate) GetPurpose(_ context.Context, _ *workflow.Dependencies, _ workflow.Workflows) otp.Purpose {
	return otp.PurposeOOBOTP
}

func (i *IntentSignupFlowStepAuthenticate) GetMessageType(_ context.Context, _ *workflow.Dependencies, workflows workflow.Workflows) otp.MessageType {
	authenticationMethod := i.authenticationMethod(workflows)
	switch authenticationMethod {
	case config.WorkflowAuthenticationMethodPrimaryOOBOTPEmail:
		return otp.MessageTypeSetupPrimaryOOB
	case config.WorkflowAuthenticationMethodPrimaryOOBOTPSMS:
		return otp.MessageTypeSetupPrimaryOOB
	case config.WorkflowAuthenticationMethodSecondaryOOBOTPEmail:
		return otp.MessageTypeSetupSecondaryOOB
	case config.WorkflowAuthenticationMethodSecondaryOOBOTPSMS:
		return otp.MessageTypeSetupSecondaryOOB
	default:
		panic(fmt.Errorf("workflow: unexpected authentication method: %v", authenticationMethod))
	}
}

var _ workflow.Intent = &IntentSignupFlowStepAuthenticate{}
var _ workflow.DataOutputer = &IntentSignupFlowStepAuthenticate{}

func (*IntentSignupFlowStepAuthenticate) Kind() string {
	return "workflowconfig.IntentSignupFlowStepAuthenticate"
}

func (*IntentSignupFlowStepAuthenticate) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	// Let the input to select which authentication method to use.
	if len(workflows.Nearest.Nodes) == 0 {
		return []workflow.Input{
			&InputTakeAuthenticationMethod{},
		}, nil
	}

	_, authenticatorCreated := FindMilestone[MilestoneDoCreateAuthenticator](workflows.Nearest)
	_, nestedStepsHandled := FindMilestone[MilestoneNestedSteps](workflows.Nearest)

	switch {
	case authenticatorCreated && !nestedStepsHandled:
		// Handle nested steps.
		return nil, nil
	default:
		return nil, workflow.ErrEOF
	}
}

func (i *IntentSignupFlowStepAuthenticate) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	current, err := signupFlowCurrent(deps, i.SignupFlow, i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	if len(workflows.Nearest.Nodes) == 0 {
		var inputTakeAuthenticationMethod inputTakeAuthenticationMethod
		if workflow.AsInput(input, &inputTakeAuthenticationMethod) &&
			// NodeCreateAuthenticatorOOBOTP sometimes does not take any input to proceed when it has target_step.
			// In that case, if the next step is also type: authenticate, then the input will be incorrectly fed to the next step.
			// To protect against this, we require the first input of each step to provide the json pointer to indicate the audience of the input.
			inputTakeAuthenticationMethod.GetJSONPointer().String() == i.JSONPointer.String() {

			authentication := inputTakeAuthenticationMethod.GetAuthenticationMethod()
			var idx int
			idx, err = i.checkAuthenticationMethod(step, authentication)
			if err != nil {
				return nil, err
			}

			switch authentication {
			case config.WorkflowAuthenticationMethodPrimaryPassword:
				fallthrough
			case config.WorkflowAuthenticationMethodSecondaryPassword:
				return workflow.NewNodeSimple(&NodeCreateAuthenticatorPassword{
					UserID:         i.UserID,
					Authentication: authentication,
				}), nil
			case config.WorkflowAuthenticationMethodPrimaryPasskey:
				// FIXME(workflow): create primary passkey
			case config.WorkflowAuthenticationMethodPrimaryOOBOTPEmail:
				fallthrough
			case config.WorkflowAuthenticationMethodSecondaryOOBOTPEmail:
				fallthrough
			case config.WorkflowAuthenticationMethodPrimaryOOBOTPSMS:
				fallthrough
			case config.WorkflowAuthenticationMethodSecondaryOOBOTPSMS:
				return workflow.NewNodeSimple(&NodeCreateAuthenticatorOOBOTP{
					SignupFlow:     i.SignupFlow,
					JSONPointer:    JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Authentication: authentication,
				}), nil
			case config.WorkflowAuthenticationMethodSecondaryTOTP:
				node, err := NewNodeCreateAuthenticatorTOTP(deps, &NodeCreateAuthenticatorTOTP{
					UserID:         i.UserID,
					Authentication: authentication,
				})
				if err != nil {
					return nil, err
				}
				return workflow.NewNodeSimple(node), nil
			}
		}
		return nil, workflow.ErrIncompatibleInput
	}

	_, authenticatorCreated := FindMilestone[MilestoneDoCreateAuthenticator](workflows.Nearest)
	_, nestedStepsHandled := FindMilestone[MilestoneNestedSteps](workflows.Nearest)

	switch {
	case authenticatorCreated && !nestedStepsHandled:
		authentication := i.authenticationMethod(workflows)
		return workflow.NewSubWorkflow(&IntentSignupFlowSteps{
			SignupFlow:  i.SignupFlow,
			JSONPointer: i.jsonPointer(step, authentication),
			UserID:      i.UserID,
		}), nil
	default:
		return nil, workflow.ErrIncompatibleInput
	}
}

func (i *IntentSignupFlowStepAuthenticate) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return map[string]interface{}{
		"json_pointer": i.JSONPointer.String(),
	}, nil
}

func (*IntentSignupFlowStepAuthenticate) step(o config.WorkflowObject) *config.WorkflowSignupFlowStep {
	step, ok := o.(*config.WorkflowSignupFlowStep)
	if !ok {
		panic(fmt.Errorf("workflow: workflow object is %T", o))
	}

	return step
}

func (*IntentSignupFlowStepAuthenticate) checkAuthenticationMethod(step *config.WorkflowSignupFlowStep, am config.WorkflowAuthenticationMethod) (idx int, err error) {
	idx = -1
	var allAllowed []config.WorkflowAuthenticationMethod

	for i, branch := range step.OneOf {
		branch := branch
		allAllowed = append(allAllowed, branch.Authentication)
		if am == branch.Authentication {
			idx = i
		}
	}

	if idx >= 0 {
		return
	}

	err = InvalidAuthenticationMethod.NewWithInfo("invalid authentication method", apierrors.Details{
		"expected": allAllowed,
		"actual":   am,
	})
	return
}

func (*IntentSignupFlowStepAuthenticate) authenticationMethod(workflows workflow.Workflows) config.WorkflowAuthenticationMethod {
	m, ok := FindMilestone[MilestoneAuthenticationMethod](workflows.Nearest)
	if !ok {
		panic(fmt.Errorf("workflow: authentication method not yet selected"))
	}

	am := m.MilestoneAuthenticationMethod()

	return am
}

func (i *IntentSignupFlowStepAuthenticate) jsonPointer(step *config.WorkflowSignupFlowStep, am config.WorkflowAuthenticationMethod) jsonpointer.T {
	for idx, branch := range step.OneOf {
		branch := branch
		if branch.Authentication == am {
			return JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("workflow: selected identification method is not allowed"))
}
