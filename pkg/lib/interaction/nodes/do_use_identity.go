package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeDoUseIdentity{})
}

type EdgeDoUseIdentity struct {
	Identity *identity.Info
}

func (e *EdgeDoUseIdentity) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	return &NodeDoUseIdentity{
		Identity: e.Identity,
	}, nil
}

type NodeDoUseIdentity struct {
	Identity *identity.Info `json:"identity"`
}

func (n *NodeDoUseIdentity) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoUseIdentity) Apply(perform func(eff interaction.Effect) error, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoUseIdentity) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
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
