package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterNode(&NodeDoResetPassword{})
}

type NodeDoResetPassword struct {
	NewPassword string `json:"new_password,omitempty"`
	Code        string `json:"code,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoResetPassword{}
var _ authflow.EffectGetter = &NodeDoResetPassword{}

func (*NodeDoResetPassword) Kind() string {
	return "NodeDoResetPassword"
}

func (n *NodeDoResetPassword) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) ([]authflow.Effect, error) {
	return []authflow.Effect{
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			milestone, ok := n.findDestination(flows)
			if ok {
				dest := milestone.MilestoneDoUseAccountRecoveryDestination()
				return deps.ResetPassword.ResetPasswordWithTarget(
					dest.TargetLoginID,
					n.Code,
					n.NewPassword,
					dest.ForgotPasswordCodeChannel(),
					dest.ForgotPasswordCodeKind(),
				)
			} else {
				// MilestoneDoUseAccountRecoveryDestination might not exist if the flow is restored
				return deps.ResetPassword.ResetPasswordByEndUser(n.Code, n.NewPassword)
			}
		}),
	}, nil
}

func (n *NodeDoResetPassword) findDestination(flows authflow.Flows) (MilestoneDoUseAccountRecoveryDestination, bool) {
	ms := authflow.FindAllMilestones[MilestoneDoUseAccountRecoveryDestination](flows.Root)
	if len(ms) == 0 {
		return nil, false
	}
	// Otherwise use the first one we find.
	return ms[0], true
}
