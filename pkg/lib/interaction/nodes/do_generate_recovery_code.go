package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeDoGenerateRecoveryCode{})
}

type EdgeDoGenerateRecoveryCode struct {
	RecoveryCodes []string
}

func (e *EdgeDoGenerateRecoveryCode) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	return &NodeDoGenerateRecoveryCode{
		RecoveryCodes: e.RecoveryCodes,
	}, nil
}

type NodeDoGenerateRecoveryCode struct {
	RecoveryCodes []string `json:"recovery_nodes"`
}

func (n *NodeDoGenerateRecoveryCode) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoGenerateRecoveryCode) Apply(perform func(eff interaction.Effect) error, graph *interaction.Graph) error {
	err := perform(interaction.EffectRun(func(ctx *interaction.Context) error {
		if len(n.RecoveryCodes) > 0 {
			_, err := ctx.MFA.ReplaceRecoveryCodes(graph.MustGetUserID(), n.RecoveryCodes)
			return err
		}

		return nil
	}))
	if err != nil {
		return err
	}

	return nil
}

func (n *NodeDoGenerateRecoveryCode) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
