package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeDoResetLockoutAttempts{})
}

type EdgeDoResetLockoutAttempts struct {
}

func (e *EdgeDoResetLockoutAttempts) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	n := &NodeDoResetLockoutAttempts{}

	return n, nil
}

type NodeDoResetLockoutAttempts struct {
}

func (n *NodeDoResetLockoutAttempts) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoResetLockoutAttempts) GetEffects() ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectOnCommit(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			authenticators := graph.GetUsedAuthenticators()
			err := ctx.Authenticators.ClearLockoutAttempts(authenticators)

			return err
		}),
	}, nil
}

func (n *NodeDoResetLockoutAttempts) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
