package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

func init() {
	newinteraction.RegisterNode(&NodeDoUseIdentity{})
}

type EdgeDoUseIdentity struct {
	Identity *identity.Info
}

func (e *EdgeDoUseIdentity) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	return &NodeDoUseIdentity{
		Identity: e.Identity,
	}, nil
}

type NodeDoUseIdentity struct {
	Identity *identity.Info `json:"identity"`
}

func (n *NodeDoUseIdentity) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeDoUseIdentity) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(ctx, graph, n)
}

func (n *NodeDoUseIdentity) UserIdentity() *identity.Info {
	return n.Identity
}

func (n *NodeDoUseIdentity) UserID() string {
	if n.Identity == nil {
		return ""
	}
	return n.Identity.UserID
}
