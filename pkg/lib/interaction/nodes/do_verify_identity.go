package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeDoVerifyIdentity{})
}

type EdgeDoVerifyIdentity struct {
	Identity         *identity.Info
	NewVerifiedClaim *verification.Claim
}

func (e *EdgeDoVerifyIdentity) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	return &NodeDoVerifyIdentity{
		Identity:         e.Identity,
		NewVerifiedClaim: e.NewVerifiedClaim,
	}, nil
}

type NodeDoVerifyIdentity struct {
	Identity         *identity.Info      `json:"identity"`
	NewVerifiedClaim *verification.Claim `json:"new_verified_claim"`
}

func (n *NodeDoVerifyIdentity) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoVerifyIdentity) GetEffects() ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectRun(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			if n.NewVerifiedClaim != nil {
				if err := ctx.Verification.MarkClaimVerified(n.NewVerifiedClaim); err != nil {
					return err
				}
			}

			return nil
		}),
	}, nil
}

func (n *NodeDoVerifyIdentity) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
