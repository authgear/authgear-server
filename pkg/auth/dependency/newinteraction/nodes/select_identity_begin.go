package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
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

type NodeSelectIdentityBegin struct {
	IdentityTypes  []authn.IdentityType   `json:"-"`
	IdentityConfig *config.IdentityConfig `json:"-"`
}

func (n *NodeSelectIdentityBegin) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	n.IdentityTypes = ctx.Config.Authentication.Identities
	n.IdentityConfig = ctx.Config.Identity
	return nil
}

func (n *NodeSelectIdentityBegin) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeSelectIdentityBegin) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	var edges []newinteraction.Edge
	for _, t := range n.IdentityTypes {
		switch t {
		case authn.IdentityTypeAnonymous:
			edges = append(edges, &EdgeUseIdentityAnonymous{
				IsCreating: false,
			})
		case authn.IdentityTypeLoginID:
			edges = append(edges, &EdgeUseIdentityLoginID{
				Mode:    UseIdentityLoginIDModeSelect,
				Configs: n.IdentityConfig.LoginID.Keys,
			})
		case authn.IdentityTypeOAuth:
			edges = append(edges, &EdgeUseIdentityOAuthProvider{
				IsCreating: false,
				Configs:    n.IdentityConfig.OAuth.Providers,
			})
		default:
			panic("interaction: unknown identity type: " + t)
		}
	}

	return edges, nil
}
