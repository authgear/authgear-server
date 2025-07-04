package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentPromoteFlowSteps{})
}

type IntentPromoteFlowSteps struct {
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`
	UserID        string                 `json:"user_id,omitempty"`
	NextStepIndex int                    `json:"next_step_index"`
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
	current, err := i.currentFlowObject(deps, flows, i)
	if err != nil {
		return nil, err
	}

	steps := current.GetSteps()

	if IsLastAuthentication(current, i.NextStepIndex) && !IsPreAuthenticatedTriggered(flows) {
		return nil, nil
	}

	if i.NextStepIndex < len(steps) {
		return nil, nil
	}

	return nil, authflow.ErrEOF
}

func (i *IntentPromoteFlowSteps) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	current, err := i.currentFlowObject(deps, flows, i)
	if err != nil {
		return nil, err
	}

	steps := current.GetSteps()
	if IsLastAuthentication(current, i.NextStepIndex) && !IsPreAuthenticatedTriggered(flows) {
		return authflow.NewSubFlow(&IntentPromoteFlowPreAuthenticated{
			FlowReference: i.FlowReference,
		}), nil
	}

	step := steps[i.NextStepIndex].(*config.AuthenticationFlowSignupFlowStep)

	// Except identify, all other steps work exactly the same as signup flow.
	var result authflow.ReactToResult
	switch step.Type {
	case config.AuthenticationFlowSignupFlowStepTypeIdentify:
		stepIdentify, err := NewIntentPromoteFlowStepIdentify(ctx, deps, flows, &IntentPromoteFlowStepIdentify{
			FlowReference: i.FlowReference,
			StepName:      step.Name,
			JSONPointer:   authflow.JSONPointerForStep(i.JSONPointer, i.NextStepIndex),
			UserID:        i.UserID,
		}, i)
		if err != nil {
			return nil, err
		}
		result = authflow.NewSubFlow(stepIdentify)
		break
	case config.AuthenticationFlowSignupFlowStepTypeVerify:
		result = authflow.NewSubFlow(&IntentSignupFlowStepVerify{
			StepName:    step.Name,
			JSONPointer: authflow.JSONPointerForStep(i.JSONPointer, i.NextStepIndex),
			UserID:      i.UserID,
		})
		break
	case config.AuthenticationFlowSignupFlowStepTypeCreateAuthenticator:
		i, err := NewIntentSignupFlowStepCreateAuthenticator(ctx, deps, flows, &IntentSignupFlowStepCreateAuthenticator{
			FlowReference: i.FlowReference,
			StepName:      step.Name,
			JSONPointer:   authflow.JSONPointerForStep(i.JSONPointer, i.NextStepIndex),
			UserID:        i.UserID,
		}, i)
		if err != nil {
			return nil, err
		}
		result = authflow.NewSubFlow(i)
		break
	case config.AuthenticationFlowSignupFlowStepTypeViewRecoveryCode:
		result = authflow.NewSubFlow(NewIntentSignupFlowStepViewRecoveryCode(ctx, deps, flows, &IntentSignupFlowStepViewRecoveryCode{
			StepName:    step.Name,
			JSONPointer: authflow.JSONPointerForStep(i.JSONPointer, i.NextStepIndex),
			UserID:      i.UserID,
		}))
		break
	case config.AuthenticationFlowSignupFlowStepTypeFillInUserProfile:
		result = authflow.NewSubFlow(&IntentSignupFlowStepFillInUserProfile{
			StepName:    step.Name,
			JSONPointer: authflow.JSONPointerForStep(i.JSONPointer, i.NextStepIndex),
			UserID:      i.UserID,
		})
		break
	case config.AuthenticationFlowSignupFlowStepTypePromptCreatePasskey:
		result = authflow.NewSubFlow(&IntentSignupFlowStepPromptCreatePasskey{
			StepName:    step.Name,
			JSONPointer: authflow.JSONPointerForStep(i.JSONPointer, i.NextStepIndex),
			UserID:      i.UserID,
		})
		break
	}

	i.NextStepIndex = i.NextStepIndex + 1
	return result, nil
}

func (i *IntentPromoteFlowSteps) currentFlowObject(deps *authflow.Dependencies, flows authflow.Flows, originNode authflow.NodeOrIntent) (config.AuthenticationFlowStepsObject, error) {
	rootObject, err := findNearestFlowObjectInFlow(deps, flows, originNode)
	if err != nil {
		return nil, err
	}
	current, err := authflow.FlowObject(rootObject, i.JSONPointer)
	if err != nil {
		return nil, err
	}
	return current.(config.AuthenticationFlowStepsObject), nil
}
