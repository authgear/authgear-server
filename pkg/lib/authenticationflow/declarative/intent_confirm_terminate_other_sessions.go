package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/uiparam"
)

func init() {
	authflow.RegisterIntent(&IntentConfirmTerminateOtherSessions{})
}

type IntentConfirmTerminateOtherSessions struct {
	UserID string `json:"user_id"`
}

var _ authflow.Intent = &IntentConfirmTerminateOtherSessions{}

func (*IntentConfirmTerminateOtherSessions) Kind() string {
	return "IntentConfirmTerminateOtherSessions"
}

func (i *IntentConfirmTerminateOtherSessions) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	if len(flows.Nearest.Nodes) == 0 {
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

	return nil, authflow.ErrEOF
}

func (i *IntentConfirmTerminateOtherSessions) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	if len(flows.Nearest.Nodes) == 0 {
		var inputConfirmTerminateOtherSessions inputConfirmTerminateOtherSessions
		if authflow.AsInput(input, &inputConfirmTerminateOtherSessions) {
			return authflow.NewNodeSimple(&NodeDidConfirmTerminateOtherSessions{}), nil
		}
	}

	return nil, authflow.ErrIncompatibleInput
}
