package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/event"
)

func init() {
	newinteraction.RegisterNode(&NodeDoCreateIdentity{})
}

type EdgeDoCreateIdentity struct {
	Identity *identity.Info
}

func (e *EdgeDoCreateIdentity) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	return &NodeDoCreateIdentity{
		Identity: e.Identity,
	}, nil
}

type NodeDoCreateIdentity struct {
	Identity *identity.Info `json:"identity"`
}

func (n *NodeDoCreateIdentity) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeDoCreateIdentity) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	err := perform(newinteraction.EffectRun(func(ctx *newinteraction.Context) error {
		if _, err := ctx.Identities.CheckDuplicated(n.Identity); err != nil {
			if errors.Is(err, identity.ErrIdentityAlreadyExists) {
				return newinteraction.ErrDuplicatedIdentity
			}
			return err
		}
		if err := ctx.Identities.Create(n.Identity); err != nil {
			return err
		}

		return nil
	}))
	if err != nil {
		return err
	}

	err = perform(newinteraction.EffectOnCommit(func(ctx *newinteraction.Context) error {
		if _, creating := graph.GetNewUserID(); creating {
			return nil
		}

		user, err := ctx.Users.Get(n.Identity.UserID)
		if err != nil {
			return err
		}

		err = ctx.Hooks.DispatchEvent(
			event.IdentityCreateEvent{
				User:     *user,
				Identity: n.Identity.ToModel(),
			},
			user,
		)
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

func (n *NodeDoCreateIdentity) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}

func (n *NodeDoCreateIdentity) UserNewIdentity() *identity.Info {
	return n.Identity
}
