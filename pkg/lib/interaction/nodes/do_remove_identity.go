package nodes

import (
	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
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
	deleteDisabled := e.Identity.DeleteDisabled(ctx.Config.Identity)
	isAdminAPI := interaction.IsAdminAPI(rawInput)
	if !isAdminAPI && deleteDisabled {
		return nil, api.ErrIdentityModifyDisabled
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
			err := ctx.Identities.Delete(n.Identity)
			if err != nil {
				return err
			}

			return nil
		}),
		interaction.EffectOnCommit(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			userRef := model.UserRef{
				Meta: model.Meta{
					ID: n.Identity.UserID,
				},
			}

			var e event.Payload
			switch n.Identity.Type {
			case model.IdentityTypeLoginID:
				loginIDType := n.Identity.LoginID.LoginIDType
				if payload, ok := nonblocking.NewIdentityLoginIDRemovedEventPayload(
					userRef,
					n.Identity.ToModel(),
					string(loginIDType),
					n.IsAdminAPI,
				); ok {
					e = payload
				}
			case model.IdentityTypeOAuth:
				e = &nonblocking.IdentityOAuthDisconnectedEventPayload{
					UserRef:  userRef,
					Identity: n.Identity.ToModel(),
					AdminAPI: n.IsAdminAPI,
				}
			case model.IdentityTypeBiometric:
				e = &nonblocking.IdentityBiometricDisabledEventPayload{
					UserRef:  userRef,
					Identity: n.Identity.ToModel(),
					AdminAPI: n.IsAdminAPI,
				}
			}

			if e != nil {
				err := ctx.Events.DispatchEventOnCommit(e)
				if err != nil {
					return err
				}
			}

			return nil
		}),
	}, nil
}

func (n *NodeDoRemoveIdentity) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
