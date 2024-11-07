package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeConfirmTerminateOtherSessionsBegin{})
}

type EdgeConfirmTerminateOtherSessionsBegin struct {
}

func (e *EdgeConfirmTerminateOtherSessionsBegin) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {

	return &NodeConfirmTerminateOtherSessionsBegin{}, nil
}

type NodeConfirmTerminateOtherSessionsBegin struct {
}

func (n *NodeConfirmTerminateOtherSessionsBegin) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeConfirmTerminateOtherSessionsBegin) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeConfirmTerminateOtherSessionsBegin) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{&EdgeConfirmTerminateOtherSessionsEnd{}}, nil
}
