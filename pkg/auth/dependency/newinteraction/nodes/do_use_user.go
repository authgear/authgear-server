package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

func init() {
	newinteraction.RegisterNode(&NodeDoUseUser{})
}

type EdgeDoUseUser struct {
	UseUserID string
}

func (e *EdgeDoUseUser) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	return &NodeDoUseUser{
		UseUserID: e.UseUserID,
	}, nil
}

type NodeDoUseUser struct {
	UseUserID string `json:"user_id"`
}

func (n *NodeDoUseUser) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeDoUseUser) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(ctx, graph, n)
}

func (n *NodeDoUseUser) UserID() string {
	return n.UseUserID
}
