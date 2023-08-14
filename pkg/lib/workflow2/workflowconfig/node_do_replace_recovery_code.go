package workflowconfig

import (
	"context"

	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterNode(&NodeDoReplaceRecoveryCode{})
}

type NodeDoReplaceRecoveryCode struct {
	UserID        string   `json:"user_id,omitempty"`
	RecoveryCodes []string `json:"recovery_codes,omitempty"`
}

var _ workflow.NodeSimple = &NodeDoReplaceRecoveryCode{}
var _ workflow.EffectGetter = &NodeDoReplaceRecoveryCode{}

func (*NodeDoReplaceRecoveryCode) Kind() string {
	return "workflowconfig.NodeDoReplaceRecoveryCode"
}

func (n *NodeDoReplaceRecoveryCode) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return []workflow.Effect{
		workflow.RunEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			_, err := deps.MFA.ReplaceRecoveryCodes(n.UserID, n.RecoveryCodes)
			return err
		}),
	}, nil
}
