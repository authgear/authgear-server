package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeDoResetLockoutAttempts{})
}

type EdgeDoResetLockoutAttempts struct {
}

func (e *EdgeDoResetLockoutAttempts) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	n := &NodeDoResetLockoutAttempts{}

	return n, nil
}

type NodeDoResetLockoutAttempts struct {
}

func (n *NodeDoResetLockoutAttempts) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoResetLockoutAttempts) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectOnCommit(func(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			userID := graph.MustGetUserID()
			methods := graph.GetUsedAuthenticationLockoutMethods()
			err := ctx.Authenticators.ClearLockoutAttempts(goCtx, userID, methods)

			return err
		}),
	}, nil
}

func (n *NodeDoResetLockoutAttempts) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(goCtx, graph, n)
}
