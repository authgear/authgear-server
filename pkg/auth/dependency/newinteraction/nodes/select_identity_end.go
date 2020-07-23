package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

type EdgeSelectIdentityEnd struct {
	RequestedIdentity *identity.Spec
	ExistingIdentity  *identity.Info
}

func (e *EdgeSelectIdentityEnd) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	return &NodeSelectIdentityEnd{
		RequestedIdentity: e.RequestedIdentity,
		ExistingIdentity:  e.ExistingIdentity,
	}, nil
}

type NodeSelectIdentityEnd struct {
	RequestedIdentity *identity.Spec `json:"requested_identity"`
	ExistingIdentity  *identity.Info `json:"existing_identity"`
}

func (n *NodeSelectIdentityEnd) Apply(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeSelectIdentityEnd) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(ctx, graph, n)
}

func (n *NodeSelectIdentityEnd) UserIdentity() *identity.Info {
	return n.ExistingIdentity
}
