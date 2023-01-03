package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeConfirmTerminateOtherSessionsBegin{})
}

type EdgeConfirmTerminateOtherSessionsBegin struct {
}

func (e *EdgeConfirmTerminateOtherSessionsBegin) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {

	return &NodeConfirmTerminateOtherSessionsBegin{}, nil
}

type NodeConfirmTerminateOtherSessionsBegin struct {
}

func (n *NodeConfirmTerminateOtherSessionsBegin) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeConfirmTerminateOtherSessionsBegin) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeConfirmTerminateOtherSessionsBegin) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{&EdgeConfirmTerminateOtherSessionsEnd{}}, nil
}
