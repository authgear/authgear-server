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

func (*IntentSignupFlowSteps) Kind() string {
	return "IntentSignupFlowSteps"
}

func (*IntentSignupFlowSteps) Milestone()            {}
func (*IntentSignupFlowSteps) MilestoneNestedSteps() {}

func (i *IntentSignupFlowSteps) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	current, err := flowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
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
	current, err := flowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
	if err != nil {
		return nil, err
	}

	steps := i.steps(current)
	nextStepIndex := len(flows.Nearest.Nodes)
	step := steps[nextStepIndex].(*config.AuthenticationFlowSignupFlowStep)

	switch step.Type {
	case config.AuthenticationFlowSignupFlowStepTypeIdentify:
		return authflow.NewSubFlow(&IntentSignupFlowStepIdentify{
			FlowReference: i.FlowReference,
			StepID:        step.ID,
			JSONPointer:   JSONPointerForStep(i.JSONPointer, nextStepIndex),
			UserID:        i.UserID,
		}), nil
	case config.AuthenticationFlowSignupFlowStepTypeVerify:
		return authflow.NewSubFlow(&IntentSignupFlowStepVerify{
			FlowReference: i.FlowReference,
			StepID:        step.ID,
			JSONPointer:   JSONPointerForStep(i.JSONPointer, nextStepIndex),
			UserID:        i.UserID,
		}), nil
	case config.AuthenticationFlowSignupFlowStepTypeAuthenticate:
		return authflow.NewSubFlow(&IntentSignupFlowStepAuthenticate{
			FlowReference: i.FlowReference,
			StepID:        step.ID,
			JSONPointer:   JSONPointerForStep(i.JSONPointer, nextStepIndex),
			UserID:        i.UserID,
		}), nil
	case config.AuthenticationFlowSignupFlowStepTypeRecoveryCode:
		return authflow.NewSubFlow(NewIntentSignupFlowStepRecoveryCode(deps, &IntentSignupFlowStepRecoveryCode{
			FlowReference: i.FlowReference,
			StepID:        step.ID,
			JSONPointer:   JSONPointerForStep(i.JSONPointer, nextStepIndex),
			UserID:        i.UserID,
		})), nil
	case config.AuthenticationFlowSignupFlowStepTypeUserProfile:
		return authflow.NewSubFlow(&IntentSignupFlowStepUserProfile{
			FlowReference: i.FlowReference,
			StepID:        step.ID,
			JSONPointer:   JSONPointerForStep(i.JSONPointer, nextStepIndex),
			UserID:        i.UserID,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (i *IntentSignupFlowSteps) steps(o config.AuthenticationFlowObject) []config.AuthenticationFlowObject {
	steps, ok := o.GetSteps()
	if !ok {
		panic(fmt.Errorf("flow object does not have steps %T", o))
	}

	return steps
}
