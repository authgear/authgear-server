package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeDoMarkClaimVerified{})
}

type NodeDoMarkClaimVerified struct {
	Claim *verification.Claim `json:"verified_claim,omitempty"`
}

func (n *NodeDoMarkClaimVerified) Kind() string {
	return "latte.NodeDoMarkClaimVerified"
}

func (n *NodeDoMarkClaimVerified) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return []workflow.Effect{
		workflow.RunEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			err := deps.Verification.MarkClaimVerified(n.Claim)
			if err != nil {
				return err
			}
			return nil
		}),
	}, nil
}

func (*NodeDoMarkClaimVerified) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeDoMarkClaimVerified) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (*NodeDoMarkClaimVerified) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
