package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/lib/api/event"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

func init() {
	newinteraction.RegisterNode(&NodeDoUpdateIdentity{})
}

type EdgeDoUpdateIdentity struct {
	IdentityBeforeUpdate *identity.Info
	IdentityAfterUpdate  *identity.Info
}

func (e *EdgeDoUpdateIdentity) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	return &NodeDoUpdateIdentity{
		IdentityBeforeUpdate: e.IdentityBeforeUpdate,
		IdentityAfterUpdate:  e.IdentityAfterUpdate,
	}, nil
}

type NodeDoUpdateIdentity struct {
	IdentityBeforeUpdate *identity.Info `json:"identity_before_update"`
	IdentityAfterUpdate  *identity.Info `json:"identity_after_update"`
}

func (n *NodeDoUpdateIdentity) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeDoUpdateIdentity) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	err := perform(newinteraction.EffectRun(func(ctx *newinteraction.Context) error {
		if _, err := ctx.Identities.CheckDuplicated(n.IdentityAfterUpdate); err != nil {
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

	err = perform(newinteraction.EffectOnCommit(func(ctx *newinteraction.Context) error {
		user, err := ctx.Users.Get(n.IdentityAfterUpdate.UserID)
		if err != nil {
			return err
		}

		err = ctx.Hooks.DispatchEvent(&event.IdentityUpdateEvent{
			User:        *user,
			OldIdentity: n.IdentityBeforeUpdate.ToModel(),
			NewIdentity: n.IdentityAfterUpdate.ToModel(),
		})
		if err != nil {
			return err
		}

		return nil
	}))
	if err != nil {
		return err
	}

	return nil
}

func (n *NodeDoUpdateIdentity) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}

func (n *NodeDoUpdateIdentity) UserIdentity() *identity.Info {
	return n.IdentityAfterUpdate
}
