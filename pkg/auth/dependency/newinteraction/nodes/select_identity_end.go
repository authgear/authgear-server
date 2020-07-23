package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

type EdgeSelectIdentityEnd struct {
	Identity    *identity.Info
	NewIdentity *identity.Info
}

func (e *EdgeSelectIdentityEnd) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	return &NodeSelectIdentityEnd{
		Identity:    e.Identity,
		NewIdentity: e.NewIdentity,
	}, nil
}

type NodeSelectIdentityEnd struct {
	Identity    *identity.Info `json:"identity"`
	NewIdentity *identity.Info `json:"new_identity"`
}

func (n *NodeSelectIdentityEnd) Apply(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	if n.NewIdentity != nil {
		panic("TODO(new_interaction): create new identity")
	}

	return nil
}

func (n *NodeSelectIdentityEnd) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(ctx, graph, n)
}

func (n *NodeSelectIdentityEnd) UserIdentity() *identity.Info {
	return n.Identity
}
