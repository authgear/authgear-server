package newinteraction

type IntentLogin struct {
	UseAnonymousUser bool `json:"use_anonymous_user"`
}

func (i *IntentLogin) DeriveFirstNode(ctx *Context, graph *Graph) (Node, error) {
	spec := EdgeSelectIdentityBegin{}
	return spec.Instantiate(ctx, graph, i)
}

func (i *IntentLogin) GetUseAnonymousUser() bool {
	return i.UseAnonymousUser
}

func (i *IntentLogin) AfterSelectIdentity(node *NodeSelectIdentityEnd) (Edge, error) {
	return &EdgeAuthenticationBegin{
		Stage:    AuthenticationStagePrimary,
		Identity: node.Identity,
	}, nil
}
