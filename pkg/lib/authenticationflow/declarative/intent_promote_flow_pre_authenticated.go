package declarative

import (
	"context"
	"fmt"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterIntent(&IntentPromoteFlowPreAuthenticated{})
}

var _ authflow.Intent = &IntentPromoteFlowPreAuthenticated{}

type IntentPromoteFlowPreAuthenticated struct {
	FlowReference authflow.FlowReference `json:"flow_reference"`
}

func (i *IntentPromoteFlowPreAuthenticated) Kind() string {
	return "IntentPromoteFlowPreAuthenticated"
}

func (i *IntentPromoteFlowPreAuthenticated) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	switch len(flows.Nearest.Nodes) {
	case 0:
		return nil, nil
	case 1:
		return nil, nil
	case 2:
		return nil, authflow.ErrEOF
	}

	panic(fmt.Errorf("unexpected node count"))
}

func (i *IntentPromoteFlowPreAuthenticated) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	switch len(flows.Nearest.Nodes) {
	case 0:
		return NewNodePreAuthenticateNodeSimple(ctx, deps, flows)
	case 1:
		subFlow, err := NewIntentSignupFlowEnforceAMRConstraints(ctx, deps, flows, &IntentSignupFlowEnforceAMRConstraintsOptions{
			FlowReference: i.FlowReference,
			UserID:        i.userID(flows),
		})
		if err != nil {
			return nil, err
		}
		return authflow.NewSubFlow(subFlow), nil
	}

	panic(fmt.Errorf("unexpected node count"))
}

func (i *IntentPromoteFlowPreAuthenticated) userID(flows authflow.Flows) string {
	userID, err := getUserID(flows)
	if err != nil {
		panic(err)
	}
	return userID
}
