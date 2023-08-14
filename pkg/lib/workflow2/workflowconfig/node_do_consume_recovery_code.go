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

var _ workflow.NodeSimple = &NodeDoConsumeRecoveryCode{}
var _ workflow.Milestone = &NodeDoConsumeRecoveryCode{}
var _ MilestoneDidAuthenticate = &NodeDoConsumeRecoveryCode{}
var _ MilestoneDidUseAuthenticationLockoutMethod = &NodeDoConsumeRecoveryCode{}
var _ workflow.EffectGetter = &NodeDoConsumeRecoveryCode{}

func (*NodeDoConsumeRecoveryCode) Kind() string {
	return "workflowconfig.NodeDoConsumeRecoveryCode"
}

func (*NodeDoConsumeRecoveryCode) Milestone()                               {}
func (*NodeDoConsumeRecoveryCode) MilestoneDidAuthenticate() (amr []string) { return }
func (*NodeDoConsumeRecoveryCode) MilestoneDidUseAuthenticationLockoutMethod() (config.AuthenticationLockoutMethod, bool) {
	return config.AuthenticationLockoutMethodRecoveryCode, true
}

func (n *NodeDoConsumeRecoveryCode) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Effect, error) {
	return []workflow.Effect{
		workflow.OnCommitEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			return deps.MFA.ConsumeRecoveryCode(n.RecoveryCode)
		}),
	}, nil
}
