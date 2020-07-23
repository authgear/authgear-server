package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterNode(&NodeSelectIdentityBegin{})
}

type InputSelectIdentityBegin interface {
	GetUseAnonymousUser() bool
}

type EdgeSelectIdentityBegin struct {
}

func (e *EdgeSelectIdentityBegin) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputSelectIdentityBegin)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	return &NodeSelectIdentityBegin{
		UseAnonymousUser: input.GetUseAnonymousUser(),
	}, nil
}

type NodeSelectIdentityBegin struct {
	UseAnonymousUser bool `json:"use_anonymous_user"`
}

func (n *NodeSelectIdentityBegin) Apply(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeSelectIdentityBegin) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	var edges []newinteraction.Edge
	for _, t := range ctx.Config.Authentication.Identities {
		switch t {
		case authn.IdentityTypeAnonymous:
			if n.UseAnonymousUser {
				// Always use anonymous user only, if requested
				return []newinteraction.Edge{&EdgeSelectIdentityAnonymous{}}, nil
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
