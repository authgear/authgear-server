package newinteraction

import "github.com/authgear/authgear-server/pkg/auth/dependency/identity"

type EdgeSelectIdentityEnd struct {
	Identity *identity.Info
}

func (e *EdgeSelectIdentityEnd) Instantiate(ctx *Context, graph *Graph, rawInput interface{}) (Node, error) {
	input, ok := rawInput.(InputSelectIdentityBegin)
	if !ok {
		return nil, ErrIncompatibleInput
	}

	return &NodeSelectIdentityBegin{
		UseAnonymousUser: input.GetUseAnonymousUser(),
	}, nil
}

type NodeSelectIdentityEnd struct {
	Identity *identity.Info `json:"identity"`
}

func (n *NodeSelectIdentityEnd) Apply(ctx *Context, graph *Graph) error {
	panic("implement me")
}

func (n *NodeSelectIdentityEnd) DeriveEdges(ctx *Context, graph *Graph) ([]Edge, error) {
	edge, err := graph.Intent.(interface {
		AfterSelectIdentity(node *NodeSelectIdentityEnd) (Edge, error)
	}).AfterSelectIdentity(n)

	return []Edge{edge}, err
}

func (n *NodeSelectIdentityEnd) UserIdentity() *identity.Info {
	return n.Identity
}
