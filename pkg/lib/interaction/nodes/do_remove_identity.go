package nodes

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
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
	isAdminAPI := false
	var adminInput interface{ IsAdminAPI() bool }
	if interaction.Input(rawInput, &adminInput) {
		isAdminAPI = adminInput.IsAdminAPI()
	}

	return &NodeDoRemoveIdentity{
		Identity:   e.Identity,
		IsAdminAPI: isAdminAPI,
	}, nil
}

type NodeDoRemoveIdentity struct {
	Identity   *identity.Info `json:"identity"`
	IsAdminAPI bool           `json:"is_admin_api"`
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

			var e event.Payload
			if n.IsAdminAPI {
				e = &nonblocking.IdentityDeletedAdminAPIRemoveIdentityEvent{
					User:     *user,
					Identity: n.Identity.ToModel(),
				}
			} else {
				e = &nonblocking.IdentityDeletedUserRemoveIdentityEvent{
					User:     *user,
					Identity: n.Identity.ToModel(),
				}
			}
			err = ctx.Hooks.DispatchEvent(e)
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
