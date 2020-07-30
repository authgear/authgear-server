package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterNode(&NodeCreateIdentityBegin{})
}

type EdgeCreateIdentityBegin struct {
	AllowAnonymousUser bool
}

func (e *EdgeCreateIdentityBegin) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	return &NodeCreateIdentityBegin{AllowAnonymousUser: e.AllowAnonymousUser}, nil
}

type NodeCreateIdentityBegin struct {
	AllowAnonymousUser bool `json:"allow_anonymous_user"`
}

func (n *NodeCreateIdentityBegin) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeCreateIdentityBegin) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	var edges []newinteraction.Edge
	for _, t := range ctx.Config.Authentication.Identities {
		switch t {
		case authn.IdentityTypeAnonymous:
			if n.AllowAnonymousUser {
				edges = append(edges, &EdgeUseIdentityAnonymous{
					IsCreating: true,
				})
			}

		case authn.IdentityTypeLoginID:
			edges = append(edges, &EdgeUseIdentityLoginID{
				IsCreating: true,
				Configs:    ctx.Config.Identity.LoginID.Keys,
			})

		case authn.IdentityTypeOAuth:
			edges = append(edges, &EdgeUseIdentityOAuthProvider{
				IsCreating: true,
				Configs:    ctx.Config.Identity.OAuth.Providers,
			})

		default:
			panic("interaction: unknown identity type: " + t)
		}
	}

	return edges, nil
}
