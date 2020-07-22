package newinteraction

type IntentLogin struct {
	UseAnonymousUser bool `json:"use_anonymous_user"`
}

func (i *IntentLogin) InstantiateRootNode(ctx *Context, graph *Graph) (Node, error) {
	spec := EdgeSelectIdentityBegin{}
	return spec.Instantiate(ctx, graph, i)
}

func (i *IntentLogin) GetUseAnonymousUser() bool {
	return i.UseAnonymousUser
}

func (i *IntentLogin) DeriveEdgesForNode(ctx *Context, graph *Graph, node Node) ([]Edge, error) {
	switch node := node.(type) {
	case *NodeSelectIdentityEnd:
		// Ensure identity exists before performing authentication
		if node.Identity == nil {
			return nil, ErrInvalidCredentials
		}

		return []Edge{
			&EdgeAuthenticationBegin{Stage: AuthenticationStagePrimary, Identity: node.Identity},
		}, nil

	default:
		panic("interaction: unexpected node")
	}
}
