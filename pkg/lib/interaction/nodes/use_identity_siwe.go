package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeUseIdentitySIWE{})
}

type InputUseIdentitySIWE interface {
	GetMessage() string
	GetSignature() string
}

type EdgeUseIdentitySIWE struct {
	IsAuthentication bool
}

func (e *EdgeUseIdentitySIWE) GetIdentityCandidates() []identity.Candidate {
	return []identity.Candidate{
		identity.NewSIWECandidate(),
	}
}

func (e *EdgeUseIdentitySIWE) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputUseIdentitySIWE
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	spec := &identity.Spec{
		Type: model.IdentityTypeSIWE,
		SIWE: &identity.SIWESpec{
			Message:   input.GetMessage(),
			Signature: input.GetSignature(),
		},
	}

	return &NodeUseIdentitySIWE{
		IsAuthentication: e.IsAuthentication,
		IdentitySpec:     spec,
	}, nil
}

type NodeUseIdentitySIWE struct {
	IsAuthentication bool           `json:"is_authentication"`
	IdentitySpec     *identity.Spec `json:"identity_spec"`
}

func (n *NodeUseIdentitySIWE) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeUseIdentitySIWE) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeUseIdentitySIWE) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{&EdgeSelectIdentityEnd{IdentitySpec: n.IdentitySpec, IsAuthentication: n.IsAuthentication}}, nil
}
