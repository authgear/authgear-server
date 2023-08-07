package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeDoConsumeRecoveryCode{})
}

type NodeDoConsumeRecoveryCode struct {
	RecoveryCode *mfa.RecoveryCode `json:"recovery_code,omitempty"`
}

var _ MilestoneAuthenticated = &NodeDoConsumeRecoveryCode{}

func (*NodeDoConsumeRecoveryCode) Milestone()              {}
func (*NodeDoConsumeRecoveryCode) MilestoneAuthenticated() {}

var _ workflow.NodeSimple = &NodeDoConsumeRecoveryCode{}

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

func (*NodeDoConsumeRecoveryCode) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeDoConsumeRecoveryCode) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, inut workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (*NodeDoConsumeRecoveryCode) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
