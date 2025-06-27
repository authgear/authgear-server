package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentLoginFlowSteps{})
}

type IntentLoginFlowSteps struct {
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`

	NextStepIndex int `json:"next_step_index"`
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

func (i *IntentLoginFlowSteps) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	current, err := i.currentFlowObject(deps, flows, i)
	if err != nil {
		return nil, err
	}

	steps := current.GetSteps()
	if IsLastAuthentication(current, i.NextStepIndex) && !IsPreAuthenticatedTriggered(flows) {
		return authflow.NewSubFlow(&IntentLoginFlowPreAuthenticated{
			FlowReference: i.FlowReference,
		}), nil
	}

	step := steps[i.NextStepIndex].(*config.AuthenticationFlowLoginFlowStep)

	var result authflow.ReactToResult
	switch step.Type {
	case config.AuthenticationFlowLoginFlowStepTypeIdentify:
		stepIdentify, err := NewIntentLoginFlowStepIdentify(ctx, deps, flows, &IntentLoginFlowStepIdentify{
			FlowReference: i.FlowReference,
			StepName:      step.Name,
			JSONPointer:   authflow.JSONPointerForStep(i.JSONPointer, i.NextStepIndex),
		}, i)
		if err != nil {
			return nil, err
		}
		result = authflow.NewSubFlow(stepIdentify)
		break
	case config.AuthenticationFlowLoginFlowStepTypeAuthenticate:
		stepAuthenticate, err := NewIntentLoginFlowStepAuthenticate(ctx, deps, flows, &IntentLoginFlowStepAuthenticate{
			FlowReference: i.FlowReference,
			StepName:      step.Name,
			JSONPointer:   authflow.JSONPointerForStep(i.JSONPointer, i.NextStepIndex),
			UserID:        i.userID(flows),
		}, i)
		if err != nil {
			return nil, err
		}
		result = authflow.NewSubFlow(stepAuthenticate)
		break
	case config.AuthenticationFlowLoginFlowStepTypeCheckAccountStatus:
		result = authflow.NewSubFlow(&IntentLoginFlowStepCheckAccountStatus{
			FlowReference: i.FlowReference,
			StepName:      step.Name,
			JSONPointer:   authflow.JSONPointerForStep(i.JSONPointer, i.NextStepIndex),
			UserID:        i.userID(flows),
		})
		break

	case config.AuthenticationFlowLoginFlowStepTypeTerminateOtherSessions:
		result = authflow.NewSubFlow(&IntentLoginFlowStepTerminateOtherSessions{
			FlowReference: i.FlowReference,
			StepName:      step.Name,
			JSONPointer:   authflow.JSONPointerForStep(i.JSONPointer, i.NextStepIndex),
			UserID:        i.userID(flows),
		})
		break
	case config.AuthenticationFlowLoginFlowStepTypeChangePassword:
		result = authflow.NewSubFlow(&IntentLoginFlowStepChangePassword{
			FlowReference: i.FlowReference,
			StepName:      step.Name,
			JSONPointer:   authflow.JSONPointerForStep(i.JSONPointer, i.NextStepIndex),
			UserID:        i.userID(flows),
		})
		break
	case config.AuthenticationFlowLoginFlowStepTypePromptCreatePasskey:
		result = authflow.NewSubFlow(&IntentLoginFlowStepPromptCreatePasskey{
			FlowReference: i.FlowReference,
			StepName:      step.Name,
			JSONPointer:   authflow.JSONPointerForStep(i.JSONPointer, i.NextStepIndex),
			UserID:        i.userID(flows),
		})
		break
	}

	i.NextStepIndex = i.NextStepIndex + 1
	return result, nil
}

func (i *IntentLoginFlowSteps) currentFlowObject(deps *authflow.Dependencies, flows authflow.Flows, origin authflow.NodeOrIntent) (config.AuthenticationFlowStepsObject, error) {
	rootObject, err := findNearestFlowObjectInFlow(deps, flows, origin)
	if err != nil {
		return nil, err
	}
	current, err := authflow.FlowObject(rootObject, i.JSONPointer)
	if err != nil {
		return nil, err
	}
	return current.(config.AuthenticationFlowStepsObject), nil
}

func (*IntentLoginFlowSteps) userID(flows authflow.Flows) string {
	userID, err := getUserID(flows)
	if err != nil {
		panic(err)
	}
	return userID
}
