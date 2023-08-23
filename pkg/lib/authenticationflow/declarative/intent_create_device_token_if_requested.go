package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterIntent(&IntentCreateDeviceTokenIfRequested{})
}

type IntentCreateDeviceTokenIfRequested struct {
	UserID string `json:"user_id,omitempty"`
}

var _ authflow.Intent = &IntentCreateDeviceTokenIfRequested{}
var _ authflow.Milestone = &IntentCreateDeviceTokenIfRequested{}
var _ MilestoneDoCreateDeviceTokenIfRequested = &IntentCreateDeviceTokenIfRequested{}

func (*IntentCreateDeviceTokenIfRequested) Kind() string {
	return "IntentCreateDeviceTokenIfRequested"
}

func (*IntentCreateDeviceTokenIfRequested) Milestone()                               {}
func (*IntentCreateDeviceTokenIfRequested) MilestoneDoCreateDeviceTokenIfRequested() {}

func (*IntentCreateDeviceTokenIfRequested) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	if len(flows.Nearest.Nodes) == 0 {
		// Take the previous input.
		return nil, nil
	}
	return nil, authflow.ErrEOF
}

func (i *IntentCreateDeviceTokenIfRequested) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	if len(flows.Nearest.Nodes) == 0 {
		var inputDeviceTokenRequested inputDeviceTokenRequested
		ok := authflow.AsInput(input, &inputDeviceTokenRequested)

		if !ok {
			// We consider this as not requested.
			// End this flow.
			return authflow.NewNodeSimple(&NodeSentinel{}), nil
		}

		requested := inputDeviceTokenRequested.GetDeviceTokenRequested()
		if !requested {
			// Simply end this flow.
			return authflow.NewNodeSimple(&NodeSentinel{}), nil
		}

		n, err := NewNodeDoCreateDeviceToken(deps, &NodeDoCreateDeviceToken{
			UserID: i.UserID,
		})
		if err != nil {
			return nil, err
		}

		return authflow.NewNodeSimple(n), nil
	}

	return nil, authflow.ErrIncompatibleInput
}
