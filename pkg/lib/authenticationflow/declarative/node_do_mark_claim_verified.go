package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
)

func init() {
	authflow.RegisterNode(&NodeDoMarkClaimVerified{})
}

type NodeDoMarkClaimVerified struct {
	Claim *verification.Claim `json:"verified_claim,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoMarkClaimVerified{}
var _ authflow.Milestone = &NodeDoMarkClaimVerified{}
var _ MilestoneDoMarkClaimVerified = &NodeDoMarkClaimVerified{}
var _ authflow.EffectGetter = &NodeDoMarkClaimVerified{}
var _ MilestoneSwitchToExistingUser = &NodeDoMarkClaimVerified{}

func (n *NodeDoMarkClaimVerified) Kind() string {
	return "latte.NodeDoMarkClaimVerified"
}

func (*NodeDoMarkClaimVerified) Milestone()                      {}
func (n *NodeDoMarkClaimVerified) MilestoneDoMarkClaimVerified() {}
func (n *NodeDoMarkClaimVerified) MilestoneSwitchToExistingUser(newUserID string) {
	// TODO(tung): Update the claim to new user id
	// but skip if it is already verified
}

func (n *NodeDoMarkClaimVerified) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			err := deps.Verification.MarkClaimVerified(n.Claim)
			if err != nil {
				return err
			}
			return nil
		}),
	}, nil
}
