package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeRemoveIdentity{})
}

type InputRemoveIdentity interface {
	GetIdentityType() authn.IdentityType
	GetIdentityID() string
}

type EdgeRemoveIdentity struct{}

func (e *EdgeRemoveIdentity) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputRemoveIdentity
	if !interaction.AsInput(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	userID := graph.MustGetUserID()
	identityType := input.GetIdentityType()
	identityID := input.GetIdentityID()

	info, err := ctx.Identities.Get(userID, identityType, identityID)
	if err != nil {
		return nil, err
	}

	return &NodeRemoveIdentity{
		IdentityInfo: info,
	}, nil
}

type NodeRemoveIdentity struct {
	IdentityInfo *identity.Info `json:"identity_info"`
}

func (n *NodeRemoveIdentity) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeRemoveIdentity) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeRemoveIdentity) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
