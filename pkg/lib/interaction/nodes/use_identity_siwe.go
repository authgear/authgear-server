package nodes

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeUseIdentitySIWE{})
}

type EdgeUseIdentitySIWE struct {
	IsAuthentication bool
}

func (e *EdgeUseIdentitySIWE) GetIdentityCandidates() []identity.Candidate {
	return []identity.Candidate{
		identity.NewSIWECandidate(),
	}
}

func (e *EdgeUseIdentitySIWE) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	spec := &identity.Spec{
		Type: model.IdentityTypePasskey,
		SIWE: &identity.SIWESpec{
			VerificationRequest: model.SIWEVerificationRequest{},
		},
	}

	return &NodeUseIdentityPasskey{
		IsAuthentication: e.IsAuthentication,
		IdentitySpec:     spec,
	}, nil
}

type NodeUseIdentitySIWE struct {
	IsAuthentication bool           `json:"is_authentication"`
	IdentitySpec     *identity.Spec `json:"identity_spec"`
}

func (n *NodeUseIdentitySIWE) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeUseIdentitySIWE) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeUseIdentitySIWE) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{&EdgeSelectIdentityEnd{IdentitySpec: n.IdentitySpec, IsAuthentication: n.IsAuthentication}}, nil
}
