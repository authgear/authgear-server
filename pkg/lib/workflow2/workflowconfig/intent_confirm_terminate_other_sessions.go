package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterIntent(&IntentConfirmTerminateOtherSessions{})
}

type IntentConfirmTerminateOtherSessions struct {
	UserID string `json:"user_id"`
}

var _ workflow.Intent = &IntentConfirmTerminateOtherSessions{}

func (*IntentConfirmTerminateOtherSessions) Kind() string {
	return "workflowconfig.IntentConfirmTerminateOtherSessions"
}

func (i *IntentConfirmTerminateOtherSessions) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (workflow.InputSchema, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		uiParam := uiparam.GetUIParam(ctx)
		clientID := uiParam.ClientID
		client, ok := deps.Config.OAuth.GetClient(clientID)
		if ok && client.MaxConcurrentSession == 1 {
			existingGrants, err := deps.OfflineGrants.ListClientOfflineGrants(clientID, i.UserID)
			if err != nil {
				return nil, err
			}

			if len(existingGrants) != 0 {
				return &InputConfirmTerminateOtherSessions{}, nil
			}
		}
	}

	return nil, workflow.ErrEOF
}

func (i *IntentConfirmTerminateOtherSessions) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		var inputConfirmTerminateOtherSessions inputConfirmTerminateOtherSessions
		if workflow.AsInput(input, &inputConfirmTerminateOtherSessions) {
			return workflow.NewNodeSimple(&NodeDidConfirmTerminateOtherSessions{}), nil
		}
	}

	return nil, workflow.ErrIncompatibleInput
}
