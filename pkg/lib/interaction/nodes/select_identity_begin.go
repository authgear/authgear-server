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

func (n *NodeSelectIdentityBegin) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeSelectIdentityBegin) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return n.deriveEdges(), nil
}

func (n *NodeSelectIdentityBegin) deriveEdges() []interaction.Edge {
	var edges []interaction.Edge
	// The checking of enable is done is the edge itself.
	// So we always add edges here.
	edges = append(edges, &EdgeUseIdentityAnonymous{})
	edges = append(edges, &EdgeUseIdentityBiometric{})

	for _, t := range n.IdentityTypes {
		switch t {
		case authn.IdentityTypeAnonymous:
			break
		case authn.IdentityTypeBiometric:
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

	// Adding EdgeIncompatibleInput to ensure graph won't end at this node
	// even no identity is configured in config file.
	edges = append(edges, &EdgeIncompatibleInput{})

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
