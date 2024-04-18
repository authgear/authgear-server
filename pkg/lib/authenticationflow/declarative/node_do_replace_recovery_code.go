package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterNode(&NodeDoReplaceRecoveryCode{})
}

type NodeDoReplaceRecoveryCode struct {
	UserID        string   `json:"user_id,omitempty"`
	RecoveryCodes []string `json:"recovery_codes,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoReplaceRecoveryCode{}
var _ authflow.EffectGetter = &NodeDoReplaceRecoveryCode{}
var _ authflow.Milestone = &NodeDoReplaceRecoveryCode{}
var _ MilestoneSwitchToExistingUser = &NodeDoReplaceRecoveryCode{}

func (*NodeDoReplaceRecoveryCode) Kind() string {
	return "NodeDoReplaceRecoveryCode"
}

func (*NodeDoReplaceRecoveryCode) Milestone() {}
func (i *NodeDoReplaceRecoveryCode) MilestoneSwitchToExistingUser(deps *authflow.Dependencies, flow *authflow.Flow, newUserID string) error {
	// This node exist only if recovery code has already been shown to user,
	// Therefore, we should still replace the code if this node was created
	i.UserID = newUserID
	return nil
}

func (n *NodeDoReplaceRecoveryCode) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			_, err := deps.MFA.ReplaceRecoveryCodes(n.UserID, n.RecoveryCodes)
			return err
		}),
	}, nil
}
