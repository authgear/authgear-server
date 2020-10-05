package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeSelectIdentityBegin{})
}

type EdgeSelectIdentityBegin struct {
	Identity *identity.Info
}

func (e *EdgeSelectIdentityBegin) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	return &NodeSelectIdentityBegin{}, nil
}

type NodeSelectIdentityBegin struct {
	IdentityTypes  []authn.IdentityType   `json:"-"`
	IdentityConfig *config.IdentityConfig `json:"-"`
}

func (n *NodeSelectIdentityBegin) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	n.IdentityTypes = ctx.Config.Authentication.Identities
	n.IdentityConfig = ctx.Config.Identity
	return nil
}

func (n *NodeSelectIdentityBegin) Apply(perform func(eff interaction.Effect) error, graph *interaction.Graph) error {
	return nil
}

func (n *NodeSelectIdentityBegin) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return n.deriveEdges(), nil
}

func (n *NodeSelectIdentityBegin) deriveEdges() []interaction.Edge {
	var edges []interaction.Edge
	// Always provide anonymous edge: checking for enabled is done in use identity node
	edges = append(edges, &EdgeUseIdentityAnonymous{
		IsCreating: false,
	})

	for _, t := range n.IdentityTypes {
		switch t {
		case authn.IdentityTypeAnonymous:
			break
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

	return edges
}

func (n *NodeSelectIdentityBegin) GetIdentityCandidates() []identity.Candidate {
	var candidates []identity.Candidate
	for _, e := range n.deriveEdges() {
		if e, ok := e.(interface{ GetIdentityCandidates() []identity.Candidate }); ok {
			candidates = append(candidates, e.GetIdentityCandidates()...)
		}
	}
	return candidates
}
