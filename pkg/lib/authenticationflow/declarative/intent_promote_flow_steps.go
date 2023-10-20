package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentPromoteFlowSteps{})
}

type IntentPromoteFlowSteps struct {
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	UserID      string        `json:"user_id,omitempty"`
}

var _ authflow.Intent = &IntentPromoteFlowSteps{}
var _ authflow.Milestone = &IntentPromoteFlowSteps{}
var _ MilestoneNestedSteps = &IntentPromoteFlowSteps{}

func (*IntentPromoteFlowSteps) Kind() string {
	return "IntentPromoteFlowSteps"
}

func (*IntentPromoteFlowSteps) Milestone()            {}
func (*IntentPromoteFlowSteps) MilestoneNestedSteps() {}

func (i *IntentPromoteFlowSteps) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
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

func (i *IntentPromoteFlowSteps) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, _ authflow.Input) (*authflow.Node, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
	if err != nil {
		return nil, err
	}

	steps := i.steps(current)
	nextStepIndex := len(flows.Nearest.Nodes)
	step := steps[nextStepIndex].(*config.AuthenticationFlowSignupFlowStep)

	// Except identify, all other steps work exactly the same as signup flow.

	switch step.Type {
	case config.AuthenticationFlowSignupFlowStepTypeIdentify:
		stepIdentify, err := NewIntentPromoteFlowStepIdentify(ctx, deps, &IntentPromoteFlowStepIdentify{
			StepName:    step.Name,
			JSONPointer: authflow.JSONPointerForStep(i.JSONPointer, nextStepIndex),
			UserID:      i.UserID,
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
			StepName:    step.Name,
			JSONPointer: authflow.JSONPointerForStep(i.JSONPointer, nextStepIndex),
			UserID:      i.UserID,
		}), nil
	case config.AuthenticationFlowSignupFlowStepTypeRecoveryCode:
		return authflow.NewSubFlow(NewIntentSignupFlowStepRecoveryCode(deps, &IntentSignupFlowStepRecoveryCode{
			StepName:    step.Name,
			JSONPointer: authflow.JSONPointerForStep(i.JSONPointer, nextStepIndex),
			UserID:      i.UserID,
		})), nil
	case config.AuthenticationFlowSignupFlowStepTypeUserProfile:
		return authflow.NewSubFlow(&IntentSignupFlowStepUserProfile{
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

func (i *IntentPromoteFlowSteps) steps(o config.AuthenticationFlowObject) []config.AuthenticationFlowObject {
	steps, ok := authflow.FlowObjectGetSteps(o)
	if !ok {
		panic(fmt.Errorf("flow object does not have steps %T", o))
	}

	return steps
}
