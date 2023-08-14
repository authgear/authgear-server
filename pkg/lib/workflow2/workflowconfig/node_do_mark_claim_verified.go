package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterNode(&NodeDoMarkClaimVerified{})
}

type NodeDoMarkClaimVerified struct {
	Claim *verification.Claim `json:"verified_claim,omitempty"`
}

var _ MilestoneDoMarkClaimVerified = &NodeDoMarkClaimVerified{}

func (*NodeDoMarkClaimVerified) Milestone()                      {}
func (n *NodeDoMarkClaimVerified) MilestoneDoMarkClaimVerified() {}

var _ workflow.NodeSimple = &NodeDoMarkClaimVerified{}
var _ workflow.EffectGetter = &NodeDoMarkClaimVerified{}

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
