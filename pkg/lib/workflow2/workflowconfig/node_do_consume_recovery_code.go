package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/config"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterNode(&NodeDoConsumeRecoveryCode{})
}

type NodeDoConsumeRecoveryCode struct {
	RecoveryCode *mfa.RecoveryCode `json:"recovery_code,omitempty"`
}

var _ Milestone = &NodeDoConsumeRecoveryCode{}

func (*NodeDoConsumeRecoveryCode) Milestone() {}

var _ MilestoneDidAuthenticate = &NodeDoConsumeRecoveryCode{}

func (*NodeDoConsumeRecoveryCode) MilestoneDidAuthenticate() (amr []string) { return }

var _ MilestoneDidUseAuthenticationLockoutMethod = &NodeDoConsumeRecoveryCode{}

func (*NodeDoConsumeRecoveryCode) MilestoneDidUseAuthenticationLockoutMethod() (config.AuthenticationLockoutMethod, bool) {
	return config.AuthenticationLockoutMethodRecoveryCode, true
}

var _ workflow.NodeSimple = &NodeDoConsumeRecoveryCode{}
var _ workflow.EffectGetter = &NodeDoConsumeRecoveryCode{}

func (*NodeDoConsumeRecoveryCode) Kind() string {
	return "workflowconfig.NodeDoConsumeRecoveryCode"
}

func (n *NodeDoConsumeRecoveryCode) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Effect, error) {
	return []workflow.Effect{
		workflow.OnCommitEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			return deps.MFA.ConsumeRecoveryCode(n.RecoveryCode)
		}),
	}, nil
}
