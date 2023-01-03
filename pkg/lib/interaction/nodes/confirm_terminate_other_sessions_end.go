package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeConfirmTerminateOtherSessionsEnd{})
}

type InputConfirmTerminateOtherSessionsEnd interface {
	GetIsConfirmed() bool
}

type EdgeConfirmTerminateOtherSessionsEnd struct {
	IsConfirmed bool
}

func (e *EdgeConfirmTerminateOtherSessionsEnd) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {

	var confirmInput InputConfirmTerminateOtherSessionsEnd
	if !interaction.Input(rawInput, &confirmInput) {
		return nil, interaction.ErrIncompatibleInput
	}

	return &NodeConfirmTerminateOtherSessionsEnd{
		IsConfirmed: confirmInput.GetIsConfirmed(),
	}, nil
}

type NodeConfirmTerminateOtherSessionsEnd struct {
	IsConfirmed bool `json:"is_confirmed"`
}

func (n *NodeConfirmTerminateOtherSessionsEnd) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeConfirmTerminateOtherSessionsEnd) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeConfirmTerminateOtherSessionsEnd) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
