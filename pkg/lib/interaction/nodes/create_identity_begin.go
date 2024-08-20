package nodes

import (
	"github.com/authgear/authgear-server/pkg/api/model"
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
	IdentityTypes         []model.IdentityType          `json:"-"`
	IdentityConfig        *config.IdentityConfig        `json:"-"`
	IdentityFeatureConfig *config.IdentityFeatureConfig `json:"-"`
}

func (n *NodeCreateIdentityBegin) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	n.IdentityTypes = ctx.Config.Authentication.Identities
	n.IdentityConfig = ctx.Config.Identity
	n.IdentityFeatureConfig = ctx.FeatureConfig.Identity
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

	// The checking of enable is done is the edge itself.
	// So we always add edges here.
	edges = append(edges, &EdgeUseIdentityBiometric{
		IsCreating: true,
	})

	for _, t := range n.IdentityTypes {
		switch t {
		case model.IdentityTypeAnonymous:
			break

		case model.IdentityTypeBiometric:
			break

		case model.IdentityTypePasskey:
			break

		case model.IdentityTypeLDAP:
			break

		case model.IdentityTypeSIWE:
			edges = append(edges, &EdgeUseIdentitySIWE{})

		case model.IdentityTypeLoginID:
			edges = append(edges, &EdgeUseIdentityLoginID{
				Mode:    UseIdentityLoginIDModeCreate,
				Configs: n.IdentityConfig.LoginID.Keys,
			})

		case model.IdentityTypeOAuth:
			edges = append(edges, &EdgeUseIdentityOAuthProvider{
				IsCreating:    true,
				Configs:       n.IdentityConfig.OAuth.Providers,
				FeatureConfig: n.IdentityFeatureConfig.OAuth.Providers,
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

func (n *NodeCreateIdentityBegin) GetIdentityCandidates() []identity.Candidate {
	var candidates []identity.Candidate
	for _, e := range n.deriveEdges() {
		if e, ok := e.(interface{ GetIdentityCandidates() []identity.Candidate }); ok {
			candidates = append(candidates, e.GetIdentityCandidates()...)
		}
	}
	return candidates
}
