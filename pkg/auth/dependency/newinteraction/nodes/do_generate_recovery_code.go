package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

func init() {
	newinteraction.RegisterNode(&NodeDoGenerateRecoveryCode{})
}

type EdgeDoGenerateRecoveryCode struct {
	RecoveryCodes []string
}

func (e *EdgeDoGenerateRecoveryCode) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	return &NodeDoGenerateRecoveryCode{
		RecoveryCodes: e.RecoveryCodes,
	}, nil
}

type NodeDoGenerateRecoveryCode struct {
	RecoveryCodes []string `json:"recovery_nodes"`
}

func (n *NodeDoGenerateRecoveryCode) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeDoGenerateRecoveryCode) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	err := perform(newinteraction.EffectRun(func(ctx *newinteraction.Context) error {
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

func (n *NodeDoGenerateRecoveryCode) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
