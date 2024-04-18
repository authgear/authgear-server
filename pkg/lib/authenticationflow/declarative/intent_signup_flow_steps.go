package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentSignupFlowSteps{})
}

type IntentSignupFlowSteps struct {
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`
	UserID        string                 `json:"user_id,omitempty"`
}

var _ authflow.Intent = &IntentSignupFlowSteps{}
var _ authflow.Milestone = &IntentSignupFlowSteps{}
var _ MilestoneNestedSteps = &IntentSignupFlowSteps{}
var _ MilestoneSwitchToExistingUser = &IntentSignupFlowSteps{}

func (*IntentSignupFlowSteps) Kind() string {
	return "IntentSignupFlowSteps"
}

func (*IntentSignupFlowSteps) Milestone()            {}
func (*IntentSignupFlowSteps) MilestoneNestedSteps() {}
func (i *IntentSignupFlowSteps) MilestoneSwitchToExistingUser(newUserID string) {
	i.UserID = newUserID
}

func (i *IntentSignupFlowSteps) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	current, err := i.currentFlowObject(deps)
	if err != nil {
		return nil, err
	}

	steps := i.steps(current)
	if len(flows.Nearest.Nodes) < len(steps) {
		return nil, nil
	}

	return nil, authflow.ErrEOF
}

func (i *IntentSignupFlowSteps) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, _ authflow.Input) (*authflow.Node, error) {
	current, err := i.currentFlowObject(deps)
	if err != nil {
		return nil, err
	}

	steps := i.steps(current)
	nextStepIndex := len(flows.Nearest.Nodes)
	step := steps[nextStepIndex].(*config.AuthenticationFlowSignupFlowStep)

	switch step.Type {
	case config.AuthenticationFlowSignupFlowStepTypeIdentify:
		stepIdentify, err := NewIntentSignupFlowStepIdentify(ctx, deps, &IntentSignupFlowStepIdentify{
			FlowReference: i.FlowReference,
			StepName:      step.Name,
			JSONPointer:   authflow.JSONPointerForStep(i.JSONPointer, nextStepIndex),
			UserID:        i.UserID,
		})
		if err != nil {
			return nil, err
		}
		return authflow.NewSubFlow(stepIdentify), nil
	case config.AuthenticationFlowSignupFlowStepTypeVerify:
		return authflow.NewSubFlow(&IntentSignupFlowStepVerify{
			StepName:    step.Name,
			JSONPointer: authflow.JSONPointerForStep(i.JSONPointer, nextStepIndex),
			UserID:      i.UserID,
		}), nil
	case config.AuthenticationFlowSignupFlowStepTypeCreateAuthenticator:
		return authflow.NewSubFlow(&IntentSignupFlowStepCreateAuthenticator{
			FlowReference: i.FlowReference,
			StepName:      step.Name,
			JSONPointer:   authflow.JSONPointerForStep(i.JSONPointer, nextStepIndex),
			UserID:        i.UserID,
		}), nil
	case config.AuthenticationFlowSignupFlowStepTypeViewRecoveryCode:
		return authflow.NewSubFlow(NewIntentSignupFlowStepViewRecoveryCode(deps, &IntentSignupFlowStepViewRecoveryCode{
			StepName:    step.Name,
			JSONPointer: authflow.JSONPointerForStep(i.JSONPointer, nextStepIndex),
			UserID:      i.UserID,
		})), nil
	case config.AuthenticationFlowSignupFlowStepTypeFillInUserProfile:
		return authflow.NewSubFlow(&IntentSignupFlowStepFillInUserProfile{
			StepName:    step.Name,
			JSONPointer: authflow.JSONPointerForStep(i.JSONPointer, nextStepIndex),
			UserID:      i.UserID,
		}), nil
	case config.AuthenticationFlowSignupFlowStepTypePromptCreatePasskey:
		return authflow.NewSubFlow(&IntentSignupFlowStepPromptCreatePasskey{
			StepName:    step.Name,
			JSONPointer: authflow.JSONPointerForStep(i.JSONPointer, nextStepIndex),
			UserID:      i.UserID,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (i *IntentSignupFlowSteps) steps(o config.AuthenticationFlowObject) []config.AuthenticationFlowObject {
	steps, ok := authflow.FlowObjectGetSteps(o)
	if !ok {
		panic(fmt.Errorf("flow object does not have steps %T", o))
	}

	return steps
}

func (i *IntentSignupFlowSteps) currentFlowObject(deps *authflow.Dependencies) (config.AuthenticationFlowObject, error) {
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
