package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

func init() {
	newinteraction.RegisterNode(&NodeGenerateRecoveryCodeEnd{})
}

type InputGenerateRecoveryCodeEnd interface {
	ViewedRecoveryCodes()
}

type EdgeGenerateRecoveryCodeEnd struct {
	RecoveryCodes []string
}

func (e *EdgeGenerateRecoveryCodeEnd) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	_, ok := rawInput.(InputGenerateRecoveryCodeEnd)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	return &NodeGenerateRecoveryCodeEnd{RecoveryCodes: e.RecoveryCodes}, nil
}

type NodeGenerateRecoveryCodeEnd struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

func (n *NodeGenerateRecoveryCodeEnd) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeGenerateRecoveryCodeEnd) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeGenerateRecoveryCodeEnd) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
