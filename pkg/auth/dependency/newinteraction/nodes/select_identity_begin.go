package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterNode(&NodeSelectIdentityBegin{})
}

type EdgeSelectIdentityBegin struct {
	Identity *identity.Info
}

func (e *EdgeSelectIdentityBegin) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	return &NodeSelectIdentityBegin{}, nil
}

type NodeSelectIdentityBegin struct{}

func (n *NodeSelectIdentityBegin) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeSelectIdentityBegin) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	var edges []newinteraction.Edge
	for _, t := range ctx.Config.Authentication.Identities {
		switch t {
		case authn.IdentityTypeAnonymous:
			edges = append(edges, &EdgeUseIdentityAnonymous{
				IsCreating: false,
			})
		case authn.IdentityTypeLoginID:
			edges = append(edges, &EdgeUseIdentityLoginID{
				Mode:    UseIdentityLoginIDModeSelect,
				Configs: ctx.Config.Identity.LoginID.Keys,
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
