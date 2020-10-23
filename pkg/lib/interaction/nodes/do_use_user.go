package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeDoUseUser{})
}

type EdgeDoUseUser struct {
	UseUserID string
}

func (e *EdgeDoUseUser) Instantiate(ctx *interaction.Context, graph *interaction.Graph, input interface{}) (interaction.Node, error) {
	return &NodeDoUseUser{
		UseUserID: e.UseUserID,
	}, nil
}

type NodeDoUseUser struct {
	UseUserID string `json:"user_id"`
}

func (n *NodeDoUseUser) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoUseUser) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeDoUseUser) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}

func (n *NodeDoUseUser) UserID() string {
	return n.UseUserID
}
