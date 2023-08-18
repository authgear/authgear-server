package workflowconfig

import (
	"context"

	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterIntent(&IntentCreateDeviceTokenIfRequested{})
}

type IntentCreateDeviceTokenIfRequested struct {
	UserID string `json:"user_id,omitempty"`
}

var _ workflow.Intent = &IntentCreateDeviceTokenIfRequested{}
var _ workflow.Milestone = &IntentCreateDeviceTokenIfRequested{}
var _ MilestoneDoCreateDeviceTokenIfRequested = &IntentCreateDeviceTokenIfRequested{}

func (*IntentCreateDeviceTokenIfRequested) Kind() string {
	return "workflowconfig.IntentCreateDeviceTokenIfRequested"
}

func (*IntentCreateDeviceTokenIfRequested) Milestone()                               {}
func (*IntentCreateDeviceTokenIfRequested) MilestoneDoCreateDeviceTokenIfRequested() {}

func (*IntentCreateDeviceTokenIfRequested) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (workflow.InputSchema, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		// Take the previous input.
		return nil, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentCreateDeviceTokenIfRequested) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		var inputDeviceTokenRequested inputDeviceTokenRequested
		ok := workflow.AsInput(input, &inputDeviceTokenRequested)

		if !ok {
			// We consider this as not requested.
			// End this workflow.
			return workflow.NewNodeSimple(&NodeSentinel{}), nil
		}

		requested := inputDeviceTokenRequested.GetDeviceTokenRequested()
		if !requested {
			// Simply end this workflow.
			return workflow.NewNodeSimple(&NodeSentinel{}), nil
		}

		n, err := NewNodeDoCreateDeviceToken(deps, &NodeDoCreateDeviceToken{
			UserID: i.UserID,
		})
		if err != nil {
			return nil, err
		}

		return workflow.NewNodeSimple(n), nil
	}

	return nil, workflow.ErrIncompatibleInput
}
