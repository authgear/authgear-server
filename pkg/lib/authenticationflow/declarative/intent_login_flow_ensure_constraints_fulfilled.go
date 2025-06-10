package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentLoginFlowEnsureConstraintsFulfilled{})
}

type IntentLoginFlowEnsureConstraintsFulfilled struct {
	FlowObject    *config.AuthenticationFlowLoginFlowStep `json:"flow_object"`
	FlowReference authenticationflow.FlowReference        `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T                           `json:"json_pointer,omitempty"`
}

func NewIntentLoginFlowEnsureConstraintsFulfilled(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, flowRef authenticationflow.FlowReference) (*IntentLoginFlowEnsureConstraintsFulfilled, error) {

	// Generate a temporary config for this step only
	flowObject := generateLoginFlowStepAuthenticateForAMRConstraints(deps.Config)

	return &IntentLoginFlowEnsureConstraintsFulfilled{
		FlowReference: flowRef,
		FlowObject:    flowObject,
		JSONPointer:   jsonpointer.T{},
	}, nil
}

var _ authenticationflow.Intent = &IntentLoginFlowEnsureConstraintsFulfilled{}
var _ authenticationflow.Milestone = &IntentLoginFlowEnsureConstraintsFulfilled{}
var _ MilestoneAuthenticationFlowObjectProvider = &IntentLoginFlowEnsureConstraintsFulfilled{}

func (*IntentLoginFlowEnsureConstraintsFulfilled) Kind() string {
	return "IntentLoginFlowEnsureConstraintsFulfilled"
}

func (i *IntentLoginFlowEnsureConstraintsFulfilled) CanReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (authenticationflow.InputSchema, error) {
	switch len(flows.Nearest.Nodes) {
	case 0:
		return nil, nil
	case 1:
		return nil, authflow.ErrEOF
	}
	panic(fmt.Errorf("unexpected number of nodes"))
}

func (i *IntentLoginFlowEnsureConstraintsFulfilled) ReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, input authenticationflow.Input) (authenticationflow.ReactToResult, error) {
	stepAuthenticate, err := NewIntentLoginFlowStepAuthenticate(ctx, deps, flows, &IntentLoginFlowStepAuthenticate{
		FlowReference: i.FlowReference,
		StepName:      "",
		JSONPointer:   i.JSONPointer,
		UserID:        i.userID(flows),
	}, i)
	if err != nil {
		return nil, err
	}
	return authflow.NewSubFlow(stepAuthenticate), nil
}

func (*IntentLoginFlowEnsureConstraintsFulfilled) userID(flows authflow.Flows) string {
	userID, err := getUserID(flows)
	if err != nil {
		panic(err)
	}
	return userID
}

func (i *IntentLoginFlowEnsureConstraintsFulfilled) Milestone() {
	return
}

// This is needed so that the child authenticate intents display a correct flow action
func (i *IntentLoginFlowEnsureConstraintsFulfilled) MilestoneAuthenticationFlowObjectProvider() config.AuthenticationFlowObject {
	return i.FlowObject
}
