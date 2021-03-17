package nodes

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeDoRemoveIdentity{})
}

type EdgeDoRemoveIdentity struct {
	Identity *identity.Info
}

func (e *EdgeDoRemoveIdentity) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	return &NodeDoRemoveIdentity{
		Identity: e.Identity,
	}, nil
}

type NodeDoRemoveIdentity struct {
	Identity *identity.Info `json:"identity"`
}

func (n *NodeDoRemoveIdentity) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoRemoveIdentity) GetEffects() ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectRun(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			userID := graph.MustGetUserID()
			remaining, err := ctx.Identities.ListByUser(userID)
			if err != nil {
				return err
			}
			remaining = identity.ApplyFilters(
				remaining,
				identity.KeepIdentifiable,
				identity.OmitID(n.Identity.ID),
			)

			if len(remaining) < 1 {
				return interaction.NewInvariantViolated(
					"RemoveLastIdentity",
					"cannot remove last identity",
					nil,
				)
			}

			err = ctx.Identities.Delete(n.Identity)
			if err != nil {
				return err
			}

			return nil
		}),
		interaction.EffectOnCommit(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			user, err := ctx.Users.Get(n.Identity.UserID)
			if err != nil {
				return err
			}

			err = ctx.Hooks.DispatchEvent(&event.IdentityDeleteEvent{
				User:     *user,
				Identity: n.Identity.ToModel(),
			})
			if err != nil {
				return err
			}

			return nil
		}),
	}, nil
}

func (n *NodeDoRemoveIdentity) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
