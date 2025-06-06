package declarative

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
)

func init() {
	authflow.RegisterIntent(&IntentLoginFlowEnsureContraintsFulfilled{})
}

type IntentLoginFlowEnsureContraintsFulfilled struct {
	FlowObject    *config.AuthenticationFlowLoginFlowStep `json:"flow_object"`
	FlowReference authenticationflow.FlowReference        `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T                           `json:"json_pointer,omitempty"`
}

func NewIntentLoginFlowEnsureContraintsFulfilled(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, flowRef authenticationflow.FlowReference) (*IntentLoginFlowEnsureContraintsFulfilled, error) {
	// TODO(tung): Traverse the flow to compute the list of OneOf
	trueValue := true
	// Generate a temporary config for this step only
	flowObject := &config.AuthenticationFlowLoginFlowStep{
		Type: config.AuthenticationFlowLoginFlowStepTypeAuthenticate,
		OneOf: []*config.AuthenticationFlowLoginFlowOneOf{
			{
				Authentication: config.AuthenticationFlowAuthenticationSecondaryPassword,
			},
			{
				Authentication: config.AuthenticationFlowAuthenticationSecondaryTOTP,
			},
			{
				Authentication: config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS,
			},
		},
		ShowUntilAMRConstraintsFulfilled: &trueValue,
	}

	return &IntentLoginFlowEnsureContraintsFulfilled{
		FlowReference: flowRef,
		FlowObject:    flowObject,
		JSONPointer:   jsonpointer.T{},
	}, nil
}

var _ authenticationflow.Intent = &IntentLoginFlowEnsureContraintsFulfilled{}
var _ authenticationflow.Milestone = &IntentLoginFlowEnsureContraintsFulfilled{}
var _ MilestoneAuthenticationFlowObjectProvider = &IntentLoginFlowEnsureContraintsFulfilled{}

func (*IntentLoginFlowEnsureContraintsFulfilled) Kind() string {
	return "IntentLoginFlowEnsureContraintsFulfilled"
}

func (i *IntentLoginFlowEnsureContraintsFulfilled) CanReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (authenticationflow.InputSchema, error) {
	switch len(flows.Nearest.Nodes) {
	case 0:
		return nil, nil
	case 1:
		return nil, authflow.ErrEOF
	}
	panic(fmt.Errorf("unexpected number of nodes"))
}

func (i *IntentLoginFlowEnsureContraintsFulfilled) ReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, input authenticationflow.Input) (authenticationflow.ReactToResult, error) {
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

func (*IntentLoginFlowEnsureContraintsFulfilled) userID(flows authflow.Flows) string {
	userID, err := getUserID(flows)
	if err != nil {
		panic(err)
	}
	return userID
}

func (i *IntentLoginFlowEnsureContraintsFulfilled) Milestone() {
	return
}

// This is needed so that the child authenticate intents display a correct flow action
func (i *IntentLoginFlowEnsureContraintsFulfilled) MilestoneAuthenticationFlowObjectProvider() config.AuthenticationFlowObject {
	return i.FlowObject
}
