package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterIntent(&IntentCreateDeviceTokenIfRequested{})
}

type IntentCreateDeviceTokenIfRequested struct {
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	UserID      string        `json:"user_id,omitempty"`
}

var _ authflow.Intent = &IntentCreateDeviceTokenIfRequested{}
var _ authflow.Milestone = &IntentCreateDeviceTokenIfRequested{}
var _ MilestoneDoCreateDeviceTokenIfRequested = &IntentCreateDeviceTokenIfRequested{}

func (*IntentCreateDeviceTokenIfRequested) Kind() string {
	return "IntentCreateDeviceTokenIfRequested"
}

func (*IntentCreateDeviceTokenIfRequested) Milestone()                               {}
func (*IntentCreateDeviceTokenIfRequested) MilestoneDoCreateDeviceTokenIfRequested() {}

func (i *IntentCreateDeviceTokenIfRequested) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	if len(flows.Nearest.Nodes) == 0 {
		flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
		if err != nil {
			return nil, err
		}
		return &InputSchemaCreateDeviceToken{
			JSONPointer:    i.JSONPointer,
			FlowRootObject: flowRootObject,
		}, nil
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
