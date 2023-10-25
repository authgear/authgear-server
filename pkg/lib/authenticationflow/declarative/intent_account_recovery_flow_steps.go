package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentAccountRecoveryFlowSteps{})
}

type IntentAccountRecoveryFlowSteps struct {
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
}

var _ authflow.Intent = &IntentAccountRecoveryFlowSteps{}
var _ authflow.Milestone = &IntentAccountRecoveryFlowSteps{}
var _ MilestoneNestedSteps = &IntentAccountRecoveryFlowSteps{}

func (*IntentAccountRecoveryFlowSteps) Kind() string {
	return "IntentAccountRecoveryFlowSteps"
}

func (*IntentAccountRecoveryFlowSteps) Milestone()            {}
func (*IntentAccountRecoveryFlowSteps) MilestoneNestedSteps() {}

func (i *IntentAccountRecoveryFlowSteps) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
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

func (i *IntentAccountRecoveryFlowSteps) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, _ authflow.Input) (*authflow.Node, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
	if err != nil {
		return nil, err
	}

	steps := i.steps(current)
	nextStepIndex := len(flows.Nearest.Nodes)
	step := steps[nextStepIndex].(*config.AuthenticationFlowAccountRecoveryFlowStep)

	switch step.Type {
	case config.AuthenticationFlowAccountRecoveryFlowTypeIdentify:
		nextStep, err := NewIntentAccountRecoveryFlowStepIdentify(ctx, deps, &IntentAccountRecoveryFlowStepIdentify{
			StepName:    step.Name,
			JSONPointer: authflow.JSONPointerForStep(i.JSONPointer, nextStepIndex),
		})
		if err != nil {
			return nil, err
		}
		return authflow.NewSubFlow(nextStep), nil
	case config.AuthenticationFlowAccountRecoveryFlowTypeSelectDestination:
		nextStep, err := NewIntentAccountRecoveryFlowStepSelectDestination(
			ctx,
			deps,
			flows,
			&IntentAccountRecoveryFlowStepSelectDestination{
				StepName:    step.Name,
				JSONPointer: authflow.JSONPointerForStep(i.JSONPointer, nextStepIndex),
			},
		)
		if err != nil {
			return nil, err
		}
		return authflow.NewSubFlow(nextStep), nil
	case config.AuthenticationFlowAccountRecoveryFlowTypeVerifyAccountRecoveryCode:
		nextStep := &IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode{
			StepName:    step.Name,
			JSONPointer: authflow.JSONPointerForStep(i.JSONPointer, nextStepIndex),
		}
		if err != nil {
			return nil, err
		}
		return authflow.NewSubFlow(nextStep), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (*IntentAccountRecoveryFlowSteps) steps(o config.AuthenticationFlowObject) []config.AuthenticationFlowObject {
	steps, ok := authflow.FlowObjectGetSteps(o)
	if !ok {
		panic(fmt.Errorf("flow object does not have steps %T", o))
	}

	return steps
}
