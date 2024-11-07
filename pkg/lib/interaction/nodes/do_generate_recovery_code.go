package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeDoGenerateRecoveryCode{})
}

type EdgeDoGenerateRecoveryCode struct {
	RecoveryCodes []string
}

func (e *EdgeDoGenerateRecoveryCode) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	return &NodeDoGenerateRecoveryCode{
		RecoveryCodes: e.RecoveryCodes,
	}, nil
}

type NodeDoGenerateRecoveryCode struct {
	RecoveryCodes []string `json:"recovery_nodes"`
}

func (n *NodeDoGenerateRecoveryCode) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoGenerateRecoveryCode) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectRun(func(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			if len(n.RecoveryCodes) > 0 {
				_, err := ctx.MFA.ReplaceRecoveryCodes(goCtx, graph.MustGetUserID(), n.RecoveryCodes)
				return err
			}

			return nil
		}),
	}, nil
}

func (n *NodeDoGenerateRecoveryCode) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(goCtx, graph, n)
}
