package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

func init() {
	newinteraction.RegisterNode(&NodeUpdateIdentityEnd{})
}

type EdgeUpdateIdentityEnd struct {
	IdentitySpec *identity.Spec
}

func (e *EdgeUpdateIdentityEnd) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
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

func (n *NodeUpdateIdentityEnd) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeUpdateIdentityEnd) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeUpdateIdentityEnd) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
