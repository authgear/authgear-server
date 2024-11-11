package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeUseIdentityPasskey{})
}

type InputUseIdentityPasskey interface {
	GetAssertionResponse() []byte
}

type EdgeUseIdentityPasskey struct {
	IsAuthentication bool
}

func (e *EdgeUseIdentityPasskey) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputUseIdentityPasskey
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	assertionResponse := input.GetAssertionResponse()
	spec := &identity.Spec{
		Type: model.IdentityTypePasskey,
		Passkey: &identity.PasskeySpec{
			AssertionResponse: assertionResponse,
		},
	}

	return &NodeUseIdentityPasskey{
		IsAuthentication: e.IsAuthentication,
		IdentitySpec:     spec,
	}, nil
}

type NodeUseIdentityPasskey struct {
	IsAuthentication bool           `json:"is_authentication"`
	IdentitySpec     *identity.Spec `json:"identity_spec"`
}

func (n *NodeUseIdentityPasskey) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeUseIdentityPasskey) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeUseIdentityPasskey) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{&EdgeSelectIdentityEnd{IdentitySpec: n.IdentitySpec, IsAuthentication: n.IsAuthentication}}, nil
}
