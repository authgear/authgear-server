package newinteraction

type InputSelectIdentityLoginID interface {
	GetLoginID() string
}

type EdgeSelectIdentityLoginID struct {
	LoginIDKey string `json:"login_id_key"`
}

func (s *EdgeSelectIdentityLoginID) Instantiate(ctx *Context, graph *Graph, rawInput interface{}) (Node, error) {
	input, ok := rawInput.(InputSelectIdentityLoginID)
	if !ok {
		return nil, ErrIncompatibleInput
	}

	return &NodeSelectIdentityLoginID{
		LoginIDKey: s.LoginIDKey,
		LoginID:    input.GetLoginID(),
	}, nil
}

type NodeSelectIdentityLoginID struct {
	LoginIDKey string `json:"login_id_key"`
	LoginID    string `json:"login_id"`
}

func (n *NodeSelectIdentityLoginID) Apply(ctx *Context, graph *Graph) error {
	panic("implement me")
}

func (n *NodeSelectIdentityLoginID) DeriveEdges(ctx *Context, graph *Graph) ([]Edge, error) {
	// return []NodeSpec{&NodeEndSelectIdentitySpec{}}, nil
	panic("implement me")
}
