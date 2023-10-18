package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
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
	return nil, nil
}

func (i *IntentPromoteFlowSteps) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, _ authflow.Input) (*authflow.Node, error) {
	// FIXME(authflow): Implement promote flow steps.
	return nil, authflow.ErrIncompatibleInput
}
