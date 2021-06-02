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
	IsAuthentication bool
}

func (e *EdgeSelectIdentityBegin) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	return &NodeSelectIdentityBegin{
		IsAuthentication: e.IsAuthentication,
	}, nil
}

type NodeSelectIdentityBegin struct {
	IsAuthentication bool                   `json:"is_authentication"`
	IdentityTypes    []authn.IdentityType   `json:"-"`
	IdentityConfig   *config.IdentityConfig `json:"-"`
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
	edges = append(edges, &EdgeUseIdentityAnonymous{
		IsAuthentication: n.IsAuthentication,
	})
	edges = append(edges, &EdgeUseIdentityBiometric{
		IsAuthentication: n.IsAuthentication,
	})

	for _, t := range n.IdentityTypes {
		switch t {
		case authn.IdentityTypeAnonymous:
			break
		case authn.IdentityTypeBiometric:
			break
		case authn.IdentityTypeLoginID:
			edges = append(edges, &EdgeUseIdentityLoginID{
				IsAuthentication: n.IsAuthentication,
				Mode:             UseIdentityLoginIDModeSelect,
				Configs:          n.IdentityConfig.LoginID.Keys,
			})
		case authn.IdentityTypeOAuth:
			edges = append(edges, &EdgeUseIdentityOAuthProvider{
				IsAuthentication: n.IsAuthentication,
				IsCreating:       false,
				Configs:          n.IdentityConfig.OAuth.Providers,
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
