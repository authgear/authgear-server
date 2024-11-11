package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeUpdateIdentityEnd{})
}

type EdgeUpdateIdentityEnd struct {
	IdentitySpec *identity.Spec
}

func (e *EdgeUpdateIdentityEnd) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	identityID := graph.MustGetUpdateIdentityID()

	oldInfo, err := ctx.Identities.Get(goCtx, identityID)
	if err != nil {
		return nil, err
	}

	if oldInfo.UserID != graph.MustGetUserID() {
		return nil, api.NewInvariantViolated(
			"IdentityNotBelongToUser",
			"identity does not belong to the user",
			nil,
		)
	}

	newInfo, err := ctx.Identities.Get(goCtx, identityID)
	if err != nil {
		return nil, err
	}

	// TODO(interaction): currently only update identity from setting page is supported
	// So LoginIDEmailByPassBlocklistAllowlist is hardcoded to be false
	// we should update to get this config from input
	// when update identity in admin portal is supported
	newInfo, err = ctx.Identities.UpdateWithSpec(goCtx, newInfo, e.IdentitySpec, identity.NewIdentityOptions{
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

func (n *NodeUpdateIdentityEnd) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeUpdateIdentityEnd) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeUpdateIdentityEnd) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(goCtx, graph, n)
}
