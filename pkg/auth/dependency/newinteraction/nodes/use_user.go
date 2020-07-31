package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

func init() {
	newinteraction.RegisterNode(&NodeUseUser{})
}

type EdgeUseUser struct {
	UseUserID string
}

func (e *EdgeUseUser) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	return &NodeUseUser{
		UseUserID: e.UseUserID,
	}, nil
}

type NodeUseUser struct {
	UseUserID string `json:"user_id"`
}

func (n *NodeUseUser) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeUseUser) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(ctx, graph, n)
}

func (n *NodeUseUser) UserID() string {
	return n.UseUserID
}
