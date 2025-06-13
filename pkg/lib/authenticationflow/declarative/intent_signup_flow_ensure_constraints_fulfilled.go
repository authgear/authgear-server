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
	authflow.RegisterIntent(&IntentSignupFlowEnsureConstraintsFulfilled{})
}

type IntentSignupFlowEnsureConstraintsFulfilled struct {
	UserID        string                                   `json:"user_id"`
	FlowObject    *config.AuthenticationFlowSignupFlowStep `json:"flow_object"`
	FlowReference authenticationflow.FlowReference         `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T                            `json:"json_pointer,omitempty"`
}

type IntentSignupFlowEnsureConstraintsFulfilledOptions struct {
	UserID        string                           `json:"user_id"`
	FlowReference authenticationflow.FlowReference `json:"flow_reference,omitempty"`
}

func NewIntentSignupFlowEnsureConstraintsFulfilled(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, opts *IntentSignupFlowEnsureConstraintsFulfilledOptions) (*IntentSignupFlowEnsureConstraintsFulfilled, error) {
	// Generate a temporary config for this step only
	flowObject := generateSignupFlowStepCreateAuthenticatorForAMRConstraints(deps.Config)

	return &IntentSignupFlowEnsureConstraintsFulfilled{
		UserID:        opts.UserID,
		FlowReference: opts.FlowReference,
		FlowObject:    flowObject,
		JSONPointer:   jsonpointer.T{},
	}, nil
}

var _ authenticationflow.Intent = &IntentSignupFlowEnsureConstraintsFulfilled{}
var _ authenticationflow.Milestone = &IntentSignupFlowEnsureConstraintsFulfilled{}
var _ MilestoneAuthenticationFlowObjectProvider = &IntentSignupFlowEnsureConstraintsFulfilled{}

func (*IntentSignupFlowEnsureConstraintsFulfilled) Kind() string {
	return "IntentSignupFlowEnsureConstraintsFulfilled"
}

func (i *IntentSignupFlowEnsureConstraintsFulfilled) CanReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (authenticationflow.InputSchema, error) {
	switch len(flows.Nearest.Nodes) {
	case 0:
		return nil, nil
	case 1:
		return nil, authflow.ErrEOF
	}
	panic(fmt.Errorf("unexpected number of nodes"))
}

func (i *IntentSignupFlowEnsureConstraintsFulfilled) ReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, input authenticationflow.Input) (authenticationflow.ReactToResult, error) {
	stepCreateAuthenticator, err := NewIntentSignupFlowStepCreateAuthenticator(ctx, deps, flows, &IntentSignupFlowStepCreateAuthenticator{
		FlowReference: i.FlowReference,
		StepName:      "",
		JSONPointer:   i.JSONPointer,
		UserID:        i.UserID,
	}, i)
	if err != nil {
		return nil, err
	}
	return authflow.NewSubFlow(stepCreateAuthenticator), nil
}

func (i *IntentSignupFlowEnsureConstraintsFulfilled) Milestone() {
	return
}

// This is needed so that the child authenticate intents display a correct flow action
func (i *IntentSignupFlowEnsureConstraintsFulfilled) MilestoneAuthenticationFlowObjectProvider() config.AuthenticationFlowObject {
	return i.FlowObject
}
