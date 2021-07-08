package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeGenerateRecoveryCodeEnd{})
}

type InputGenerateRecoveryCodeEnd interface {
	ViewedRecoveryCodes()
}

type EdgeGenerateRecoveryCodeEnd struct {
	RecoveryCodes []string
}

func (e *EdgeGenerateRecoveryCodeEnd) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputGenerateRecoveryCodeEnd
	if !interaction.AsInput(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	return &NodeGenerateRecoveryCodeEnd{RecoveryCodes: e.RecoveryCodes}, nil
}

type NodeGenerateRecoveryCodeEnd struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

func (n *NodeGenerateRecoveryCodeEnd) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeGenerateRecoveryCodeEnd) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeGenerateRecoveryCodeEnd) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
