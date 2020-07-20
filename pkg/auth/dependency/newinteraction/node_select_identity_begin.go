package newinteraction

type InputSelectIdentityBegin interface {
	GetUseAnonymousUser() bool
}

type EdgeSelectIdentityBegin struct {
}

func (e *EdgeSelectIdentityBegin) Instantiate(ctx *Context, graph *Graph, rawInput interface{}) (Node, error) {
	input, ok := rawInput.(InputSelectIdentityBegin)
	if !ok {
		return nil, ErrIncompatibleInput
	}

	return &NodeSelectIdentityBegin{
		UseAnonymousUser: input.GetUseAnonymousUser(),
	}, nil
}

type NodeSelectIdentityBegin struct {
	UseAnonymousUser bool `json:"use_anonymous_user"`
}

func (n *NodeSelectIdentityBegin) Apply(ctx *Context, graph *Graph) error {
	panic("implement me")
}

func (n *NodeSelectIdentityBegin) DeriveEdges(ctx *Context, graph *Graph) ([]Edge, error) {
	panic("implement me")
}
