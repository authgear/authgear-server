package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterIntent(&IntentLoginFlowStepCheckAccountStatus{})
}

type IntentLoginFlowStepCheckAccountStatus struct {
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	StepName      string                 `json:"step_name,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`
	UserID        string                 `json:"user_id,omitempty"`
}

var _ authflow.Intent = &IntentLoginFlowStepCheckAccountStatus{}

func (*IntentLoginFlowStepCheckAccountStatus) Kind() string {
	return "IntentLoginFlowStepCheckAccountStatus"
}

func (i *IntentLoginFlowStepCheckAccountStatus) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	// This step does not require any input.
	if len(flows.Nearest.Nodes) == 0 {
		return nil, nil
	}

	return nil, authflow.ErrEOF
}

func (i *IntentLoginFlowStepCheckAccountStatus) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, _ authflow.Input) (*authflow.Node, error) {
	u, err := deps.Users.GetRaw(i.UserID)
	if err != nil {
		return nil, err
	}

	err = u.AccountStatus().Check()
	if err != nil {
		return nil, err
	}

	return authflow.NewNodeSimple(&NodeSentinel{}), nil
}
