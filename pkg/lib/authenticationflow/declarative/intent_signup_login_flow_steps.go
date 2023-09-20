package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentSignupLoginFlowSteps{})
}

type IntentSignupLoginFlowSteps struct {
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
}

var _ authflow.Intent = &IntentSignupLoginFlowSteps{}
var _ authflow.Milestone = &IntentSignupLoginFlowSteps{}
var _ MilestoneNestedSteps = &IntentSignupLoginFlowSteps{}

func (*IntentSignupLoginFlowSteps) Kind() string {
	return "IntentSignupLoginFlowSteps"
}

func (*IntentSignupLoginFlowSteps) Milestone()            {}
func (*IntentSignupLoginFlowSteps) MilestoneNestedSteps() {}

func (i *IntentSignupLoginFlowSteps) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
	if err != nil {
		return nil, err
	}

	steps := i.steps(current)
	if len(flows.Nearest.Nodes) < len(steps) {
		return nil, nil
	}

	return nil, authflow.ErrEOF
}

func (i *IntentSignupLoginFlowSteps) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, _ authflow.Input) (*authflow.Node, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
	if err != nil {
		return nil, err
	}

	steps := i.steps(current)
	nextStepIndex := len(flows.Nearest.Nodes)
	step := steps[nextStepIndex].(*config.AuthenticationFlowSignupLoginFlowStep)

	switch step.Type {
	case config.AuthenticationFlowSignupLoginFlowStepTypeIdentify:
		stepIdentify, err := NewIntentSignupLoginFlowStepIdentify(ctx, deps, &IntentSignupLoginFlowStepIdentify{
			StepName:    step.Name,
			JSONPointer: authflow.JSONPointerForStep(i.JSONPointer, nextStepIndex),
		})
		if err != nil {
			return nil, err
		}
		return authflow.NewSubFlow(stepIdentify), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (*IntentSignupLoginFlowSteps) steps(o config.AuthenticationFlowObject) []config.AuthenticationFlowObject {
	steps, ok := authflow.FlowObjectGetSteps(o)
	if !ok {
		panic(fmt.Errorf("flow object does not have steps %T", o))
	}

	return steps
}
