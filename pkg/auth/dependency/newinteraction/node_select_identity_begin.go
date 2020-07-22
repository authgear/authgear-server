package newinteraction

import "github.com/authgear/authgear-server/pkg/core/authn"

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
	return nil
}

func (n *NodeSelectIdentityBegin) DeriveEdges(ctx *Context, graph *Graph) ([]Edge, error) {
	var edges []Edge
	for _, t := range ctx.Config.Authentication.Identities {
		switch t {
		case authn.IdentityTypeAnonymous:
			if n.UseAnonymousUser {
				// Always use anonymous user only, if requested
				return []Edge{&EdgeSelectIdentityAnonymous{}}, nil
			}

		case authn.IdentityTypeLoginID:
			for _, c := range ctx.Config.Identity.LoginID.Keys {
				edges = append(edges, &EdgeSelectIdentityLoginID{Config: c})
			}

		case authn.IdentityTypeOAuth:
			for _, c := range ctx.Config.Identity.OAuth.Providers {
				edges = append(edges, &EdgeSelectIdentityOAuth{Config: c})
			}

		default:
			panic("interaction: unknown identity type: " + t)
		}
	}

	return edges, nil
}
