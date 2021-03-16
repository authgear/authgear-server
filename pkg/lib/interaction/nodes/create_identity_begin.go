package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeCreateIdentityBegin{})
}

type EdgeCreateIdentityBegin struct{}

func (e *EdgeCreateIdentityBegin) Instantiate(ctx *interaction.Context, graph *interaction.Graph, input interface{}) (interaction.Node, error) {
	return &NodeCreateIdentityBegin{}, nil
}

type NodeCreateIdentityBegin struct {
	IdentityTypes  []authn.IdentityType   `json:"-"`
	IdentityConfig *config.IdentityConfig `json:"-"`
}

func (n *NodeCreateIdentityBegin) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	n.IdentityTypes = ctx.Config.Authentication.Identities
	n.IdentityConfig = ctx.Config.Identity
	return nil
}

func (n *NodeCreateIdentityBegin) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeCreateIdentityBegin) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return n.deriveEdges(), nil
}

func (n *NodeCreateIdentityBegin) deriveEdges() []interaction.Edge {
	var edges []interaction.Edge
	for _, t := range n.IdentityTypes {
		switch t {
		case authn.IdentityTypeAnonymous:
			break

		case authn.IdentityTypeLoginID:
			edges = append(edges, &EdgeUseIdentityLoginID{
				Mode:    UseIdentityLoginIDModeCreate,
				Configs: n.IdentityConfig.LoginID.Keys,
			})

		case authn.IdentityTypeOAuth:
			edges = append(edges, &EdgeUseIdentityOAuthProvider{
				IsCreating: true,
				Configs:    n.IdentityConfig.OAuth.Providers,
			})

		default:
			panic("interaction: unknown identity type: " + t)
		}
	}

	return edges
}

func (n *NodeCreateIdentityBegin) GetIdentityCandidates() []identity.Candidate {
	var candidates []identity.Candidate
	for _, e := range n.deriveEdges() {
		if e, ok := e.(interface{ GetIdentityCandidates() []identity.Candidate }); ok {
			candidates = append(candidates, e.GetIdentityCandidates()...)
		}
	}
	return candidates
}
