package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type IntentSignupFlowStepCreateAuthenticatorTarget interface {
	GetOOBOTPClaims(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (map[model.ClaimName]string, error)
}

func init() {
	authflow.RegisterIntent(&IntentSignupFlowStepCreateAuthenticator{})
}

type IntentSignupFlowStepCreateAuthenticatorData struct {
	TypedData
	Options []CreateAuthenticatorOption `json:"options,omitempty"`
}

func NewIntentSignupFlowStepCreateAuthenticatorData(d IntentSignupFlowStepCreateAuthenticatorData) IntentSignupFlowStepCreateAuthenticatorData {
	d.Type = DataTypeCreateAuthenticatorData
	return d
}

var _ authflow.Data = &IntentSignupFlowStepCreateAuthenticatorData{}

func (m IntentSignupFlowStepCreateAuthenticatorData) Data() {}

type IntentSignupFlowStepCreateAuthenticator struct {
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`
	StepName      string                 `json:"step_name,omitempty"`
	UserID        string                 `json:"user_id,omitempty"`
}

var _ authflow.TargetStep = &IntentSignupFlowStepCreateAuthenticator{}

func (i *IntentSignupFlowStepCreateAuthenticator) GetName() string {
	return i.StepName
}

func (i *IntentSignupFlowStepCreateAuthenticator) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ authflow.Intent = &IntentSignupFlowStepCreateAuthenticator{}
var _ authflow.DataOutputer = &IntentSignupFlowStepCreateAuthenticator{}
var _ authflow.Milestone = &IntentSignupFlowStepCreateAuthenticator{}
var _ MilestoneSwitchToExistingUser = &IntentSignupFlowStepCreateAuthenticator{}

func (*IntentSignupFlowStepCreateAuthenticator) Milestone() {}
func (i *IntentSignupFlowStepCreateAuthenticator) MilestoneSwitchToExistingUser(newUserID string) {
	i.UserID = newUserID
}

func (*IntentSignupFlowStepCreateAuthenticator) Kind() string {
	return "IntentSignupFlowStepCreateAuthenticator"
}

func (i *IntentSignupFlowStepCreateAuthenticator) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	// Let the input to select which authentication method to use.
	if len(flows.Nearest.Nodes) == 0 {
		flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
		if err != nil {
			return nil, err
		}
		current, err := authflow.FlowObject(flowRootObject, i.JSONPointer)
		if err != nil {
			return nil, err
		}
		step := i.step(current)
		return &InputSchemaSignupFlowStepCreateAuthenticator{
			FlowRootObject: flowRootObject,
			JSONPointer:    i.JSONPointer,
			OneOf:          step.OneOf,
		}, nil
	}

	_, authenticatorCreated := authflow.FindMilestone[MilestoneDoCreateAuthenticator](flows.Nearest)
	_, nestedStepsHandled := authflow.FindMilestone[MilestoneNestedSteps](flows.Nearest)

	switch {
	case authenticatorCreated && !nestedStepsHandled:
		// Handle nested steps.
		return nil, nil
	default:
		return nil, authflow.ErrEOF
	}
}

func (i *IntentSignupFlowStepCreateAuthenticator) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	current, err := i.currentFlowObject(deps)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	if len(flows.Nearest.Nodes) == 0 {
		var inputTakeAuthenticationMethod inputTakeAuthenticationMethod
		if authflow.AsInput(input, &inputTakeAuthenticationMethod) {

			authentication := inputTakeAuthenticationMethod.GetAuthenticationMethod()
			idx, err := i.checkAuthenticationMethod(deps, step, authentication)
			if err != nil {
				return nil, err
			}

			switch authentication {
			case config.AuthenticationFlowAuthenticationPrimaryPassword:
				fallthrough
			case config.AuthenticationFlowAuthenticationSecondaryPassword:
				return authflow.NewNodeSimple(&NodeCreateAuthenticatorPassword{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Authentication: authentication,
				}), nil
			case config.AuthenticationFlowAuthenticationPrimaryPasskey:
				// Cannot create passkey in this step.
				return nil, authflow.ErrIncompatibleInput
			case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
				fallthrough
			case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
				fallthrough
			case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
				fallthrough
			case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
				return authflow.NewSubFlow(&IntentCreateAuthenticatorOOBOTP{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Authentication: authentication,
				}), nil
			case config.AuthenticationFlowAuthenticationSecondaryTOTP:
				node, err := NewNodeCreateAuthenticatorTOTP(deps, &NodeCreateAuthenticatorTOTP{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Authentication: authentication,
				})
				if err != nil {
					return nil, err
				}
				return authflow.NewNodeSimple(node), nil
			}
		}
		return nil, authflow.ErrIncompatibleInput
	}

	_, authenticatorCreated := authflow.FindMilestone[MilestoneDoCreateAuthenticator](flows.Nearest)
	_, nestedStepsHandled := authflow.FindMilestone[MilestoneNestedSteps](flows.Nearest)

	switch {
	case authenticatorCreated && !nestedStepsHandled:
		authentication := i.authenticationMethod(flows)
		return authflow.NewSubFlow(&IntentSignupFlowSteps{
			FlowReference: i.FlowReference,
			JSONPointer:   i.jsonPointer(step, authentication),
			UserID:        i.UserID,
		}), nil
	default:
		return nil, authflow.ErrIncompatibleInput
	}
}

func (i *IntentSignupFlowStepCreateAuthenticator) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	current, err := i.currentFlowObject(deps)
	if err != nil {
		return nil, err
	}
	step := i.step(current)
	options, err := NewCreateAuthenticationOptions(ctx, deps, flows, step, i.UserID)
	if err != nil {
		return nil, err
	}

	return NewIntentSignupFlowStepCreateAuthenticatorData(IntentSignupFlowStepCreateAuthenticatorData{
		Options: options,
	}), nil
}

func (*IntentSignupFlowStepCreateAuthenticator) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowSignupFlowStep {
	step, ok := o.(*config.AuthenticationFlowSignupFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}

func (i *IntentSignupFlowStepCreateAuthenticator) checkAuthenticationMethod(deps *authflow.Dependencies, step *config.AuthenticationFlowSignupFlowStep, am config.AuthenticationFlowAuthentication) (idx int, err error) {
	idx = -1

	for index, branch := range step.OneOf {
		branch := branch
		if am == branch.Authentication {
			idx = index
		}
	}

	if idx >= 0 {
		return
	}

	err = authflow.ErrIncompatibleInput
	return
}

func (*IntentSignupFlowStepCreateAuthenticator) authenticationMethod(flows authflow.Flows) config.AuthenticationFlowAuthentication {
	m, ok := authflow.FindMilestone[MilestoneAuthenticationMethod](flows.Nearest)
	if !ok {
		panic(fmt.Errorf("authentication method not yet selected"))
	}

	am := m.MilestoneAuthenticationMethod()

	return am
}

func (i *IntentSignupFlowStepCreateAuthenticator) jsonPointer(step *config.AuthenticationFlowSignupFlowStep, am config.AuthenticationFlowAuthentication) jsonpointer.T {
	for idx, branch := range step.OneOf {
		branch := branch
		if branch.Authentication == am {
			return authflow.JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("selected identification method is not allowed"))
}

func (i *IntentSignupFlowStepCreateAuthenticator) currentFlowObject(deps *authflow.Dependencies) (config.AuthenticationFlowObject, error) {
	rootObject, err := flowRootObject(deps, i.FlowReference)
	if err != nil {
		return nil, err
	}
	current, err := authflow.FlowObject(rootObject, i.JSONPointer)
	if err != nil {
		return nil, err
	}
	return current, nil
}
