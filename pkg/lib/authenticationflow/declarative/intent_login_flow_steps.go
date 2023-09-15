package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentLoginFlowSteps{})
}

type IntentLoginFlowSteps struct {
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
}

var _ authflow.Intent = &IntentLoginFlowSteps{}
var _ authflow.Milestone = &IntentLoginFlowSteps{}
var _ MilestoneNestedSteps = &IntentLoginFlowSteps{}

func (*IntentLoginFlowSteps) Kind() string {
	return "IntentLoginFlowSteps"
}

func (*IntentLoginFlowSteps) Milestone()            {}
func (*IntentLoginFlowSteps) MilestoneNestedSteps() {}

func (i *IntentLoginFlowSteps) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
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

func (i *IntentLoginFlowSteps) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, _ authflow.Input) (*authflow.Node, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
	if err != nil {
		return nil, err
	}

	steps := i.steps(current)
	nextStepIndex := len(flows.Nearest.Nodes)
	step := steps[nextStepIndex].(*config.AuthenticationFlowLoginFlowStep)

	switch step.Type {
	case config.AuthenticationFlowLoginFlowStepTypeIdentify:
		stepIdentify, err := NewIntentLoginFlowStepIdentify(ctx, deps, &IntentLoginFlowStepIdentify{
			StepID:      step.ID,
			JSONPointer: authflow.JSONPointerForStep(i.JSONPointer, nextStepIndex),
		})
		if err != nil {
			return nil, err
		}
		return authflow.NewSubFlow(stepIdentify), nil
	case config.AuthenticationFlowLoginFlowStepTypeAuthenticate:
		stepAuthenticate, err := NewIntentLoginFlowStepAuthenticate(ctx, deps, flows, &IntentLoginFlowStepAuthenticate{
			StepID:      step.ID,
			JSONPointer: authflow.JSONPointerForStep(i.JSONPointer, nextStepIndex),
			UserID:      i.userID(flows),
		})
		if err != nil {
			return nil, err
		}
		return authflow.NewSubFlow(stepAuthenticate), nil
	case config.AuthenticationFlowLoginFlowStepTypeChangePassword:
		return authflow.NewSubFlow(&IntentLoginFlowStepChangePassword{
			StepID:      step.ID,
			JSONPointer: authflow.JSONPointerForStep(i.JSONPointer, nextStepIndex),
			UserID:      i.userID(flows),
		}), nil
	case config.AuthenticationFlowLoginFlowStepTypePromptCreatePasskey:
		return authflow.NewSubFlow(&IntentLoginFlowStepPromptCreatePasskey{
			StepID:      step.ID,
			JSONPointer: authflow.JSONPointerForStep(i.JSONPointer, nextStepIndex),
			UserID:      i.userID(flows),
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (*IntentLoginFlowSteps) steps(o config.AuthenticationFlowObject) []config.AuthenticationFlowObject {
	steps, ok := authflow.FlowObjectGetSteps(o)
	if !ok {
		panic(fmt.Errorf("flow object does not have steps %T", o))
	}

	return steps
}

func (*IntentLoginFlowSteps) userID(flows authflow.Flows) string {
	userID, err := getUserID(flows)
	if err != nil {
		panic(err)
	}
	return userID
}
