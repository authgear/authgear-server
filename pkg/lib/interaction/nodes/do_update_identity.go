package nodes

import (
	"errors"

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

func (e *EdgeDoUpdateIdentity) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	isAdminAPI := interaction.IsAdminAPI(rawInput)
	modifyDisabled := e.IdentityBeforeUpdate.ModifyDisabled(ctx.Config.Identity)
	if !isAdminAPI && modifyDisabled {
		return nil, interaction.ErrIdentityModifyDisabled
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

func (n *NodeDoUpdateIdentity) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoUpdateIdentity) GetEffects() ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectRun(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			if _, err := ctx.Identities.CheckDuplicated(n.IdentityAfterUpdate); err != nil {
				if errors.Is(err, identity.ErrIdentityAlreadyExists) {
					s1 := n.IdentityBeforeUpdate.ToSpec()
					s2 := n.IdentityAfterUpdate.ToSpec()
					return identityFillDetails(interaction.ErrDuplicatedIdentity, &s2, &s1)
				}
				return err
			}

			if err := ctx.Identities.Update(n.IdentityAfterUpdate); err != nil {
				return err
			}

			return nil
		}),
		interaction.EffectOnCommit(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			userRef := model.UserRef{
				Meta: model.Meta{
					ID: n.IdentityAfterUpdate.UserID,
				},
			}

			var e event.Payload
			switch n.IdentityAfterUpdate.Type {
			case model.IdentityTypeLoginID:
				loginIDType := n.IdentityAfterUpdate.LoginID.LoginIDType
				e = nonblocking.NewIdentityLoginIDUpdatedEventPayload(
					userRef,
					n.IdentityAfterUpdate.ToModel(),
					n.IdentityBeforeUpdate.ToModel(),
					string(loginIDType),
					n.IsAdminAPI,
				)
			}

			if e != nil {
				err := ctx.Events.DispatchEvent(e)
				if err != nil {
					return err
				}
			}

			return nil
		}),
	}, nil
}

func (n *NodeDoUpdateIdentity) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}

func (n *NodeDoUpdateIdentity) UserIdentity() *identity.Info {
	return n.IdentityAfterUpdate
}
