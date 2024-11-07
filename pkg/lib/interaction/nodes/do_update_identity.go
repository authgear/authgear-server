package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeDoUpdateIdentity{})
}

type EdgeDoUpdateIdentity struct {
	IdentityBeforeUpdate *identity.Info
	IdentityAfterUpdate  *identity.Info
}

func (e *EdgeDoUpdateIdentity) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	isAdminAPI := interaction.IsAdminAPI(rawInput)
	updateDisabled := e.IdentityBeforeUpdate.UpdateDisabled(ctx.Config.Identity)
	if !isAdminAPI && updateDisabled {
		return nil, api.ErrIdentityModifyDisabled
	}
	return &NodeDoUpdateIdentity{
		IdentityBeforeUpdate: e.IdentityBeforeUpdate,
		IdentityAfterUpdate:  e.IdentityAfterUpdate,
		IsAdminAPI:           isAdminAPI,
	}, nil
}

type NodeDoUpdateIdentity struct {
	IdentityBeforeUpdate *identity.Info `json:"identity_before_update"`
	IdentityAfterUpdate  *identity.Info `json:"identity_after_update"`
	IsAdminAPI           bool           `json:"is_admin_api"`
}

func (n *NodeDoUpdateIdentity) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoUpdateIdentity) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectRun(func(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			if _, err := ctx.Identities.CheckDuplicated(goCtx, n.IdentityAfterUpdate); err != nil {
				if identity.IsErrDuplicatedIdentity(err) {
					s1 := n.IdentityBeforeUpdate.ToSpec()
					s2 := n.IdentityAfterUpdate.ToSpec()
					return identity.NewErrDuplicatedIdentity(&s2, &s1)
				}
				return err
			}

			if err := ctx.Identities.Update(goCtx, n.IdentityBeforeUpdate, n.IdentityAfterUpdate); err != nil {
				return err
			}

			return nil
		}),
		interaction.EffectOnCommit(func(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			userRef := model.UserRef{
				Meta: model.Meta{
					ID: n.IdentityAfterUpdate.UserID,
				},
			}

			var e event.Payload
			switch n.IdentityAfterUpdate.Type {
			case model.IdentityTypeLoginID:
				loginIDType := n.IdentityAfterUpdate.LoginID.LoginIDType
				if payload, ok := nonblocking.NewIdentityLoginIDUpdatedEventPayload(
					userRef,
					n.IdentityAfterUpdate.ToModel(),
					n.IdentityBeforeUpdate.ToModel(),
					string(loginIDType),
					n.IsAdminAPI,
				); ok {
					e = payload
				}
			}

			if e != nil {
				err := ctx.Events.DispatchEventOnCommit(goCtx, e)
				if err != nil {
					return err
				}
			}

			return nil
		}),
	}, nil
}

func (n *NodeDoUpdateIdentity) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(goCtx, graph, n)
}

func (n *NodeDoUpdateIdentity) UserIdentity() *identity.Info {
	return n.IdentityAfterUpdate
}
