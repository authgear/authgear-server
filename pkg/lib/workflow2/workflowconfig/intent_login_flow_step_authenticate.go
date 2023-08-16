package workflowconfig

import (
	"context"
	"fmt"
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

type IntentLoginFlowStepAuthenticateTarget interface {
	GetIdentity(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) *identity.Info
}

func init() {
	workflow.RegisterIntent(&IntentLoginFlowStepAuthenticate{})
}

type IntentLoginFlowStepAuthenticateData struct {
	Candidates []UseAuthenticationCandidate `json:"candidates"`
}

var _ workflow.Data = IntentLoginFlowStepAuthenticateData{}

func (m IntentLoginFlowStepAuthenticateData) Data() {}

type IntentLoginFlowStepAuthenticate struct {
	LoginFlow   string        `json:"login_flow,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	StepID      string        `json:"step_id,omitempty"`
	UserID      string        `json:"user_id,omitempty"`
}

var _ WorkflowStep = &IntentLoginFlowStepAuthenticate{}

func (i *IntentLoginFlowStepAuthenticate) GetID() string {
	return i.StepID
}

func (i *IntentLoginFlowStepAuthenticate) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ IntentLoginFlowStepChangePasswordTarget = &IntentLoginFlowStepAuthenticate{}

func (*IntentLoginFlowStepAuthenticate) GetPasswordAuthenticator(_ context.Context, _ *workflow.Dependencies, workflows workflow.Workflows) (info *authenticator.Info, ok bool) {
	m, ok := workflow.FindMilestone[MilestoneDidVerifyAuthenticator](workflows.Nearest)
	if !ok {
		return
	}

	ok = false
	n := m.MilestoneDidVerifyAuthenticator()

	if n.Authenticator.Type == model.AuthenticatorTypePassword {
		if n.PasswordChangeRequired {
			info = n.Authenticator
			ok = true
		}
	}

	return
}

var _ workflow.Intent = &IntentLoginFlowStepAuthenticate{}
var _ workflow.Boundary = &IntentLoginFlowStepAuthenticate{}
var _ workflow.DataOutputer = &IntentLoginFlowStepAuthenticate{}

func NewIntentLoginFlowStepAuthenticate(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, i *IntentLoginFlowStepAuthenticate) (*IntentLoginFlowStepAuthenticate, error) {
	// OutputData will include usable authenticators.
	// If it returns error, there is no usable authenticators.
	// This intent cannot proceed if there is no usable authenticators.
	// Therefore, we prevent from adding this intent to the workflow if such case happens.
	_, err := i.OutputData(ctx, deps, workflows)
	if err != nil {
		return nil, err
	}

	return i, nil
}

func (*IntentLoginFlowStepAuthenticate) Kind() string {
	return "workflowconfig.IntentLoginFlowStepAuthenticate"
}

func (i *IntentLoginFlowStepAuthenticate) Boundary() string {
	return i.JSONPointer.String()
}

func (i *IntentLoginFlowStepAuthenticate) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	current, err := loginFlowCurrent(deps, i.LoginFlow, i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	deviceTokenEnabled := i.deviceTokenEnabled(step)

	_, deviceTokenInspected := workflow.FindMilestone[MilestoneDeviceTokenInspected](workflows.Nearest)

	_, authenticationMethodSelected := workflow.FindMilestone[MilestoneAuthenticationMethod](workflows.Nearest)

	_, authenticated := workflow.FindMilestone[MilestoneDidAuthenticate](workflows.Nearest)

	_, deviceTokenCreatedIfRequested := workflow.FindMilestone[MilestoneDoCreateDeviceTokenIfRequested](workflows.Nearest)

	_, nestedStepsHandled := workflow.FindMilestone[MilestoneNestedSteps](workflows.Nearest)

	switch {
	case deviceTokenEnabled && !deviceTokenInspected:
		// Inspect the device token
		return nil, nil
	case !authenticationMethodSelected:
		// Let the input to select which authentication method to use.
		return []workflow.Input{
			&InputTakeAuthenticationMethod{},
		}, nil
	case !authenticated:
		// This branch is only reached when there is a programming error.
		// We expect the selected authentication method to be authenticated before this intent becomes input reactor again.
		panic(fmt.Errorf("workflow: unauthenticated"))

	case deviceTokenEnabled && !deviceTokenCreatedIfRequested:
		// We look at the current input to see if device token is request.
		// So we do not need to take another input.
		return nil, nil
	case !nestedStepsHandled:
		// Handle nested steps.
		return nil, nil
	default:
		return nil, workflow.ErrEOF
	}
}

func (i *IntentLoginFlowStepAuthenticate) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	current, err := loginFlowCurrent(deps, i.LoginFlow, i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	deviceTokenEnabled := i.deviceTokenEnabled(step)

	_, deviceTokenInspected := workflow.FindMilestone[MilestoneDeviceTokenInspected](workflows.Nearest)

	_, authenticationMethodSelected := workflow.FindMilestone[MilestoneAuthenticationMethod](workflows.Nearest)

	_, authenticated := workflow.FindMilestone[MilestoneDidAuthenticate](workflows.Nearest)

	_, deviceTokenCreatedIfRequested := workflow.FindMilestone[MilestoneDoCreateDeviceTokenIfRequested](workflows.Nearest)

	_, nestedStepsHandled := workflow.FindMilestone[MilestoneNestedSteps](workflows.Nearest)

	switch {
	case deviceTokenEnabled && !deviceTokenInspected:
		return workflow.NewSubWorkflow(&IntentInspectDeviceToken{
			UserID: i.UserID,
		}), nil
	case !authenticationMethodSelected:
		var inputTakeAuthenticationMethod inputTakeAuthenticationMethod
		if workflow.AsInput(input, &inputTakeAuthenticationMethod) {
			authentication := inputTakeAuthenticationMethod.GetAuthenticationMethod()
			var idx int

			candidates, err := getAuthenticationCandidatesForStep(ctx, deps, workflows, i.UserID, step)
			if err != nil {
				return nil, err
			}

			_, err = i.checkAuthenticationMethod(step, candidates, authentication)
			if err != nil {
				return nil, err
			}

			switch authentication {
			case config.WorkflowAuthenticationMethodPrimaryPassword:
				fallthrough
			case config.WorkflowAuthenticationMethodSecondaryPassword:
				return workflow.NewNodeSimple(&NodeUseAuthenticatorPassword{
					UserID:         i.UserID,
					Authentication: authentication,
				}), nil
			case config.WorkflowAuthenticationMethodPrimaryPasskey:
				// FIXME(workflow): authenticate with passkey
			case config.WorkflowAuthenticationMethodPrimaryOOBOTPEmail:
				fallthrough
			case config.WorkflowAuthenticationMethodSecondaryOOBOTPEmail:
				fallthrough
			case config.WorkflowAuthenticationMethodPrimaryOOBOTPSMS:
				fallthrough
			case config.WorkflowAuthenticationMethodSecondaryOOBOTPSMS:
				return workflow.NewSubWorkflow(&IntentUseAuthenticatorOOBOTP{
					LoginFlow:         i.LoginFlow,
					JSONPointer:       JSONPointerForOneOf(i.JSONPointer, idx),
					JSONPointerToStep: i.JSONPointer,
					UserID:            i.UserID,
					Authentication:    authentication,
				}), nil
			case config.WorkflowAuthenticationMethodSecondaryTOTP:
				return workflow.NewNodeSimple(&NodeUseAuthenticatorTOTP{
					UserID:         i.UserID,
					Authentication: authentication,
				}), nil
			case config.WorkflowAuthenticationMethodRecoveryCode:
				return workflow.NewNodeSimple(&NodeUseRecoveryCode{
					UserID:         i.UserID,
					Authentication: authentication,
				}), nil
			case config.WorkflowAuthenticationMethodDeviceToken:
				// Device token is handled transparently.
				return nil, workflow.ErrIncompatibleInput
			}
		}

		return nil, workflow.ErrIncompatibleInput
	case !authenticated:
		panic(fmt.Errorf("workflow: unauthenticated"))
	case deviceTokenEnabled && !deviceTokenCreatedIfRequested:
		return workflow.NewSubWorkflow(&IntentCreateDeviceTokenIfRequested{
			UserID: i.UserID,
		}), nil
	case !nestedStepsHandled:
		authentication := i.authenticationMethod(workflows)
		return workflow.NewSubWorkflow(&IntentLoginFlowSteps{
			LoginFlow:   i.LoginFlow,
			JSONPointer: i.jsonPointer(step, authentication),
		}), nil
	default:
		return nil, workflow.ErrIncompatibleInput
	}
}

func (i *IntentLoginFlowStepAuthenticate) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (workflow.Data, error) {
	current, err := loginFlowCurrent(deps, i.LoginFlow, i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	candidates, err := getAuthenticationCandidatesForStep(ctx, deps, workflows, i.UserID, step)
	if err != nil {
		return nil, err
	}

	return IntentLoginFlowStepAuthenticateData{
		Candidates: candidates,
	}, nil
}

func (i *IntentLoginFlowStepAuthenticate) checkAuthenticationMethod(step *config.WorkflowLoginFlowStep, candidates []UseAuthenticationCandidate, am config.WorkflowAuthenticationMethod) (idx int, err error) {
	idx = -1

	allAllowed := i.getAllAllowed(step)

	for i := range allAllowed {
		thisMethod := allAllowed[i]
		for _, candidate := range candidates {
			if thisMethod == candidate.AuthenticationMethod && thisMethod == am {
				idx = i
			}
		}
	}

	if idx >= 0 {
		return
	}

	err = InvalidAuthenticationMethod.New("invalid authentication method")
	return
}

func (*IntentLoginFlowStepAuthenticate) getAllAllowed(step *config.WorkflowLoginFlowStep) []config.WorkflowAuthenticationMethod {
	// Make empty slice.
	allAllowed := []config.WorkflowAuthenticationMethod{}

	for _, branch := range step.OneOf {
		branch := branch
		allAllowed = append(allAllowed, branch.Authentication)
	}

	return allAllowed
}

func (i *IntentLoginFlowStepAuthenticate) deviceTokenEnabled(step *config.WorkflowLoginFlowStep) bool {
	allAllowed := i.getAllAllowed(step)
	for _, am := range allAllowed {
		if am == config.WorkflowAuthenticationMethodDeviceToken {
			return true
		}
	}
	return false
}

func (*IntentLoginFlowStepAuthenticate) step(o config.WorkflowObject) *config.WorkflowLoginFlowStep {
	step, ok := o.(*config.WorkflowLoginFlowStep)
	if !ok {
		panic(fmt.Errorf("workflow: workflow object is %T", o))
	}

	return step
}

func (*IntentLoginFlowStepAuthenticate) authenticationMethod(workflows workflow.Workflows) config.WorkflowAuthenticationMethod {
	m, ok := workflow.FindMilestone[MilestoneAuthenticationMethod](workflows.Nearest)
	if !ok {
		panic(fmt.Errorf("workflow: authentication method not yet selected"))
	}

	am := m.MilestoneAuthenticationMethod()

	return am
}

func (i *IntentLoginFlowStepAuthenticate) jsonPointer(step *config.WorkflowLoginFlowStep, am config.WorkflowAuthenticationMethod) jsonpointer.T {
	for idx, branch := range step.OneOf {
		branch := branch
		if branch.Authentication == am {
			return JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("workflow: selected authentication method is not allowed"))
}
