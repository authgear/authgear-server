package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterNode(&NodeCreateIdentityBegin{})
}

type InputCreateIdentityBegin interface {
}

type EdgeCreateIdentityBegin struct {
}

func (e *EdgeCreateIdentityBegin) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	return &NodeCreateIdentityBegin{}, nil
}

type NodeCreateIdentityBegin struct {
}

func (n *NodeCreateIdentityBegin) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeCreateIdentityBegin) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	var edges []newinteraction.Edge
	for _, t := range ctx.Config.Authentication.Identities {
		switch t {
		case authn.IdentityTypeAnonymous:
			panic("TODO(interaction): handle anonymous signup")
		case authn.IdentityTypeLoginID:
			edges = append(edges, &EdgeUseIdentityLoginID{
				IsCreating: false,
				Configs:    ctx.Config.Identity.LoginID.Keys,
			})

		case authn.IdentityTypeOAuth:
			edges = append(edges, &EdgeUseIdentityOAuthProvider{
				IsCreating: false,
				Configs:    ctx.Config.Identity.OAuth.Providers,
			})
		default:
			panic("interaction: unknown identity type: " + t)
		}
	}

	return edges, nil
}
