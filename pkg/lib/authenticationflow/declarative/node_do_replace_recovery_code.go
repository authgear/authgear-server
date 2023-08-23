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

func (*NodeDoReplaceRecoveryCode) Kind() string {
	return "NodeDoReplaceRecoveryCode"
}

func (n *NodeDoReplaceRecoveryCode) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			_, err := deps.MFA.ReplaceRecoveryCodes(n.UserID, n.RecoveryCodes)
			return err
		}),
	}, nil
}
