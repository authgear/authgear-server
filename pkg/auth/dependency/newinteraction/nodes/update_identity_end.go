package nodes

import (
	"errors"

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
		IdentitySpec:         e.IdentitySpec,
		IdentityBeforeUpdate: oldInfo,
		IdentityAfterUpdate:  newInfo,
	}, nil
}

type NodeUpdateIdentityEnd struct {
	IdentitySpec         *identity.Spec `json:"identity_spec"`
	IdentityBeforeUpdate *identity.Info `json:"identity_before_update"`
	IdentityAfterUpdate  *identity.Info `json:"identity_after_update"`
}

func (n *NodeUpdateIdentityEnd) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	err := perform(newinteraction.EffectRun(func(ctx *newinteraction.Context) error {
		if err := ctx.Identities.CheckDuplicated(n.IdentityAfterUpdate); err != nil {
			if errors.Is(err, identity.ErrIdentityAlreadyExists) {
				return newinteraction.ErrDuplicatedIdentity
			}
			return err
		}

		if err := ctx.Identities.Update(n.IdentityAfterUpdate); err != nil {
			return err
		}

		return nil
	}))
	if err != nil {
		return err
	}

	// TODO(interaction): dispatch identity update event if not creating user

	return nil
}

func (n *NodeUpdateIdentityEnd) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(ctx, graph, n)
}

func (n *NodeUpdateIdentityEnd) UserIdentity() *identity.Info {
	return n.IdentityAfterUpdate
}
