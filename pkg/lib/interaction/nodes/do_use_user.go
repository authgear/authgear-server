package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeDoUseUser{})
}

type EdgeDoUseUser struct {
	UseUserID string
}

func (e *EdgeDoUseUser) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, input interface{}) (interaction.Node, error) {
	return &NodeDoUseUser{
		UseUserID: e.UseUserID,
	}, nil
}

type NodeDoUseUser struct {
	UseUserID string `json:"user_id"`
}

func (n *NodeDoUseUser) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoUseUser) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeDoUseUser) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(goCtx, graph, n)
}

func (n *NodeDoUseUser) UserID() string {
	return n.UseUserID
}
