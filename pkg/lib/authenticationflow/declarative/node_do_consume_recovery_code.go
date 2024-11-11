package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterNode(&NodeDoConsumeRecoveryCode{})
}

type NodeDoConsumeRecoveryCode struct {
	RecoveryCode *mfa.RecoveryCode `json:"recovery_code,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoConsumeRecoveryCode{}
var _ authflow.Milestone = &NodeDoConsumeRecoveryCode{}
var _ MilestoneDidAuthenticate = &NodeDoConsumeRecoveryCode{}
var _ MilestoneDidUseAuthenticationLockoutMethod = &NodeDoConsumeRecoveryCode{}
var _ authflow.EffectGetter = &NodeDoConsumeRecoveryCode{}

func (*NodeDoConsumeRecoveryCode) Kind() string {
	return "NodeDoConsumeRecoveryCode"
}

func (*NodeDoConsumeRecoveryCode) Milestone()                               {}
func (*NodeDoConsumeRecoveryCode) MilestoneDidAuthenticate() (amr []string) { return }
func (*NodeDoConsumeRecoveryCode) MilestoneDidUseAuthenticationLockoutMethod() (config.AuthenticationLockoutMethod, bool) {
	return config.AuthenticationLockoutMethodRecoveryCode, true
}

func (n *NodeDoConsumeRecoveryCode) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) ([]authflow.Effect, error) {
	return []authflow.Effect{
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			return deps.MFA.ConsumeRecoveryCode(ctx, n.RecoveryCode)
		}),
	}, nil
}
