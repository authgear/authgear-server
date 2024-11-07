package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeVerifiedIdentity{})
}

type NodeVerifiedIdentity struct {
	IdentityID       string              `json:"identity_id"`
	NewVerifiedClaim *verification.Claim `json:"verified_claim,omitempty"`
}

func (n *NodeVerifiedIdentity) Kind() string {
	return "latte.NodeVerifiedIdentity"
}

func (n *NodeVerifiedIdentity) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return []workflow.Effect{
		workflow.RunEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			if n.NewVerifiedClaim == nil {
				// Verified already; skip marking
				return nil
			}

			if err := deps.Verification.MarkClaimVerified(ctx, n.NewVerifiedClaim); err != nil {
				return err
			}
			return nil
		}),
	}, nil
}

func (*NodeVerifiedIdentity) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeVerifiedIdentity) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (*NodeVerifiedIdentity) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
