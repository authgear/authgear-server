package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentReauthFlowSteps{})
}

type IntentReauthFlowSteps struct {
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`
	NextStepIndex int                    `json:"next_step_index"`
}

var _ authflow.Intent = &IntentReauthFlowSteps{}
var _ authflow.Milestone = &IntentReauthFlowSteps{}
var _ MilestoneNestedSteps = &IntentReauthFlowSteps{}

func (*IntentReauthFlowSteps) Kind() string {
	return "IntentReauthFlowSteps"
}

func (*IntentReauthFlowSteps) Milestone()            {}
func (*IntentReauthFlowSteps) MilestoneNestedSteps() {}

func (i *IntentReauthFlowSteps) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
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

func (i *IntentReauthFlowSteps) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	current, err := i.currentFlowObject(deps, flows, i)
	if err != nil {
		return nil, err
	}

	steps := current.GetSteps()
	if IsLastAuthentication(current, i.NextStepIndex) && !IsPreAuthenticatedTriggered(flows) {
		return authflow.NewSubFlow(&IntentReauthFlowPreAuthenticated{
			FlowReference: i.FlowReference,
		}), nil
	}

	step := steps[i.NextStepIndex].(*config.AuthenticationFlowReauthFlowStep)

	var result authflow.ReactToResult
	switch step.Type {
	case config.AuthenticationFlowReauthFlowStepTypeIdentify:
		stepIdentify, err := NewIntentReauthFlowStepIdentify(ctx, deps, flows, &IntentReauthFlowStepIdentify{
			FlowReference: i.FlowReference,
			StepName:      step.Name,
			JSONPointer:   authflow.JSONPointerForStep(i.JSONPointer, i.NextStepIndex),
		}, i)
		if err != nil {
			return nil, err
		}
		result = authflow.NewSubFlow(stepIdentify)
		break
	case config.AuthenticationFlowReauthFlowStepTypeAuthenticate:
		stepAuthenticate, err := NewIntentReauthFlowStepAuthenticate(ctx, deps, flows, &IntentReauthFlowStepAuthenticate{
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
	}

	i.NextStepIndex = i.NextStepIndex + 1
	return result, nil
}

func (i *IntentReauthFlowSteps) currentFlowObject(deps *authflow.Dependencies, flows authflow.Flows, origin authflow.NodeOrIntent) (config.AuthenticationFlowStepsObject, error) {
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

func (*IntentReauthFlowSteps) userID(flows authflow.Flows) string {
	userID, err := getUserID(flows)
	if err != nil {
		panic(err)
	}
	return userID
}
