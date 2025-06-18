package declarative

import (
	"context"
	"fmt"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterIntent(&IntentReauthFlowPreAuthenticated{})
}

var _ authflow.Intent = &IntentReauthFlowPreAuthenticated{}

type IntentReauthFlowPreAuthenticated struct {
	FlowReference authflow.FlowReference `json:"flow_reference"`
}

func (i *IntentReauthFlowPreAuthenticated) Kind() string {
	return "IntentReauthFlowPreAuthenticated"
}

func (i *IntentReauthFlowPreAuthenticated) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
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

func (i *IntentReauthFlowPreAuthenticated) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	switch len(flows.Nearest.Nodes) {
	case 0:
		return NewNodePreAuthenticateNodeSimple(ctx, deps, flows)
	case 1:
		subFlow, err := NewIntentReauthFlowEnforceAMRConstraints(ctx, deps, flows, i.FlowReference)
		if err != nil {
			return nil, err
		}
		return authflow.NewSubFlow(subFlow), nil
	}

	panic(fmt.Errorf("unexpected node count"))
}
