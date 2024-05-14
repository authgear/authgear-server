package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/uiparam"
)

func init() {
	authflow.RegisterIntent(&IntentLoginFlowStepTerminateOtherSessions{})
}

type IntentLoginFlowStepTerminateOtherSessions struct {
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	StepName      string                 `json:"step_name,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`
	UserID        string                 `json:"user_id,omitempty"`
}

var _ authflow.Intent = &IntentLoginFlowStepTerminateOtherSessions{}

func (*IntentLoginFlowStepTerminateOtherSessions) Kind() string {
	return "IntentLoginFlowStepTerminateOtherSessions"
}

func (i *IntentLoginFlowStepTerminateOtherSessions) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	// Find out whether this step is needed.
	if len(flows.Nearest.Nodes) == 0 {
		return nil, nil
	}

	return nil, authflow.ErrEOF
}

func (i *IntentLoginFlowStepTerminateOtherSessions) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	uiParam := uiparam.GetUIParam(ctx)
	clientID := uiParam.ClientID
	client, ok := deps.Config.OAuth.GetClient(clientID)
	if ok && client.MaxConcurrentSession == 1 {
		existingGrants, err := deps.OfflineGrants.ListClientOfflineGrants(clientID, i.UserID)
		if err != nil {
			return nil, err
		}

		if len(existingGrants) != 0 {
			return authflow.NewNodeSimple(&NodeLoginFlowTerminateOtherSessions{
				JSONPointer: i.JSONPointer,
			}), nil
		}
	}

	return authflow.NewNodeSimple(&NodeSentinel{}), nil
}
