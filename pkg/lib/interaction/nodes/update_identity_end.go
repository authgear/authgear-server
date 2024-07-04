package nodes

import (
	"fmt"

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
	identityID := graph.MustGetUpdateIdentityID()

	oldInfo, err := ctx.Identities.Get(identityID)
	if err != nil {
		return nil, err
	}

	if oldInfo.UserID != graph.MustGetUserID() {
		return nil, fmt.Errorf("identity does not belong to the user")
	}

	newInfo, err := ctx.Identities.Get(identityID)
	if err != nil {
		return nil, err
	}

	// TODO(interaction): currently only update identity from setting page is supported
	// So LoginIDEmailByPassBlocklistAllowlist is hardcoded to be false
	// we should update to get this config from input
	// when update identity in admin portal is supported
	newInfo, err = ctx.Identities.UpdateWithSpec(newInfo, e.IdentitySpec, identity.NewIdentityOptions{
		LoginIDEmailByPassBlocklistAllowlist: false,
	})
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

func (n *NodeUpdateIdentityEnd) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeUpdateIdentityEnd) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
