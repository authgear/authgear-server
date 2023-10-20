package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentRequestAccountRecoveryFlowSteps{})
}

type IntentRequestAccountRecoveryFlowSteps struct {
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
}

var _ authflow.Intent = &IntentRequestAccountRecoveryFlowSteps{}
var _ authflow.Milestone = &IntentRequestAccountRecoveryFlowSteps{}
var _ MilestoneNestedSteps = &IntentRequestAccountRecoveryFlowSteps{}

func (*IntentRequestAccountRecoveryFlowSteps) Kind() string {
	return "IntentRequestAccountRecoveryFlowSteps"
}

func (*IntentRequestAccountRecoveryFlowSteps) Milestone()            {}
func (*IntentRequestAccountRecoveryFlowSteps) MilestoneNestedSteps() {}

func (i *IntentRequestAccountRecoveryFlowSteps) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
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

func (i *IntentRequestAccountRecoveryFlowSteps) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, _ authflow.Input) (*authflow.Node, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
	if err != nil {
		return nil, err
	}

	steps := i.steps(current)
	nextStepIndex := len(flows.Nearest.Nodes)
	step := steps[nextStepIndex].(*config.AuthenticationFlowRequestAccountRecoveryFlowStep)

	switch step.Type {
	case config.AuthenticationFlowAccountRecoveryFlowTypeIdentify:
		stepIdentify, err := NewIntentRequestAccountRecoveryFlowStepIdentify(ctx, deps, &IntentRequestAccountRecoveryFlowStepIdentify{
			StepName:    step.Name,
			JSONPointer: authflow.JSONPointerForStep(i.JSONPointer, nextStepIndex),
		})
		if err != nil {
			return nil, err
		}
		return authflow.NewSubFlow(stepIdentify), nil
	case config.AuthenticationFlowAccountRecoveryFlowTypeSelectDestination:
		// FIXME(tung)
		return nil, nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (*IntentRequestAccountRecoveryFlowSteps) steps(o config.AuthenticationFlowObject) []config.AuthenticationFlowObject {
	steps, ok := authflow.FlowObjectGetSteps(o)
	if !ok {
		panic(fmt.Errorf("flow object does not have steps %T", o))
	}

	return steps
}
