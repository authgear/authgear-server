package declarative

import (
	"context"
	"fmt"
	"strconv"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentAccountRecoveryFlowSteps{})
}

type IntentAccountRecoveryFlowSteps struct {
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`
	StartFrom     jsonpointer.T          `json:"start_from,omitempty"`
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

func (i *IntentAccountRecoveryFlowSteps) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, _ authflow.Input) (*authflow.Node, error) {
	current, err := i.currentFlowObject(deps)
	if err != nil {
		return nil, err
	}

	steps := i.steps(current)
	nextStepIndex := i.nextStepIndex(flows)

	if i.initialStepIndex() > nextStepIndex {
		// fast forward by inserting NodeSentinel
		return authflow.NewNodeSimple(&NodeSentinel{}), nil
	}

	step := steps[nextStepIndex].(*config.AuthenticationFlowAccountRecoveryFlowStep)

	switch step.Type {
	case config.AuthenticationFlowAccountRecoveryFlowTypeIdentify:
		nextStep, err := NewIntentAccountRecoveryFlowStepIdentify(ctx, deps, &IntentAccountRecoveryFlowStepIdentify{
			FlowReference: i.FlowReference,
			StepName:      step.Name,
			JSONPointer:   authflow.JSONPointerForStep(i.JSONPointer, nextStepIndex),
			StartFrom:     i.StartFrom,
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
				FlowReference: i.FlowReference,
				StepName:      step.Name,
				JSONPointer:   authflow.JSONPointerForStep(i.JSONPointer, nextStepIndex),
			},
		)
		if err != nil {
			return nil, err
		}
		return authflow.NewSubFlow(nextStep), nil
	case config.AuthenticationFlowAccountRecoveryFlowTypeVerifyAccountRecoveryCode:
		nextStep := &IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode{
			StepName:      step.Name,
			JSONPointer:   authflow.JSONPointerForStep(i.JSONPointer, nextStepIndex),
			FlowReference: i.FlowReference,
			StartFrom:     i.StartFrom,
		}
		return authflow.NewSubFlow(nextStep), nil
	case config.AuthenticationFlowAccountRecoveryFlowTypeResetPassword:
		nextStep := &IntentAccountRecoveryFlowStepResetPassword{
			StepName:    step.Name,
			JSONPointer: authflow.JSONPointerForStep(i.JSONPointer, nextStepIndex),
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

func (i *IntentAccountRecoveryFlowSteps) currentFlowObject(deps *authflow.Dependencies) (config.AuthenticationFlowObject, error) {
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

func (i *IntentAccountRecoveryFlowSteps) initialStepIndex() int {
	startFrom := authflow.JSONPointerSubtract(i.StartFrom, i.JSONPointer)
	if len(startFrom) < 2 || startFrom[0] != authflow.JsonPointerTokenSteps {
		return 0
	}
	currentIdx, err := strconv.Atoi(startFrom[1])
	if err != nil {
		return 0
	}
	return currentIdx
}

func (i *IntentAccountRecoveryFlowSteps) nextStepIndex(flows authflow.Flows) int {
	return len(flows.Nearest.Nodes)
}
