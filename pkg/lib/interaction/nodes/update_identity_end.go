package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeUpdateIdentityEnd{})
}

type EdgeUpdateIdentityEnd struct {
	IdentitySpec *identity.Spec
}

func (e *EdgeUpdateIdentityEnd) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	userID := graph.MustGetUserID()
	identityType := e.IdentitySpec.Type
	identityID := graph.MustGetUpdateIdentityID()

	oldInfo, err := ctx.Identities.Get(userID, identityType, identityID)
	if err != nil {
		return nil, err
	}

	newInfo, err := ctx.Identities.Get(userID, identityType, identityID)
	if err != nil {
		return nil, err
	}

	newInfo, err = ctx.Identities.UpdateWithSpec(newInfo, e.IdentitySpec)
	if err != nil {
		return nil, err
	}

	return &NodeUpdateIdentityEnd{
		IdentityBeforeUpdate: oldInfo,
		IdentityAfterUpdate:  newInfo,
	}, nil
}

type NodeUpdateIdentityEnd struct {
	IdentityBeforeUpdate *identity.Info `json:"identity_before_update"`
	IdentityAfterUpdate  *identity.Info `json:"identity_after_update"`
}

func (n *NodeUpdateIdentityEnd) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeUpdateIdentityEnd) Apply(perform func(eff interaction.Effect) error, graph *interaction.Graph) error {
	return nil
}

func (n *NodeUpdateIdentityEnd) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
