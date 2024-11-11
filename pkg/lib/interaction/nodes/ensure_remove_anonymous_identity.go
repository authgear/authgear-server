package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeEnsureRemoveAnonymousIdentity{})
}

type EdgeEnsureRemoveAnonymousIdentity struct{}

func (e *EdgeEnsureRemoveAnonymousIdentity) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, input interface{}) (interaction.Node, error) {
	return &NodeEnsureRemoveAnonymousIdentity{
		IsAdminAPI: interaction.IsAdminAPI(input),
	}, nil
}

type NodeEnsureRemoveAnonymousIdentity struct {
	AnonymousIdentity *identity.Info `json:"anonymous_identity,omitempty"`
	IsAdminAPI        bool           `json:"is_admin_api"`
}

func (n *NodeEnsureRemoveAnonymousIdentity) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	userID := graph.MustGetUserID()
	iis, err := ctx.Identities.ListByUser(goCtx, graph.MustGetUserID())
	if err != nil {
		return err
	}
	anonymousIdentities := identity.ApplyFilters(
		iis,
		identity.KeepType(model.IdentityTypeAnonymous),
	)
	if len(anonymousIdentities) > 1 {
		panic("interaction: more than 1 anonymous identities: " + userID)
	}
	if len(anonymousIdentities) == 1 {
		if !n.IsAdminAPI {
			// This node is used when adding new identity
			// when admin add a new identifiable identity to the anonymous user
			// anonymous will be removed
			panic("interaction: unexpected anonymous user adding identity")
		}
		n.AnonymousIdentity = anonymousIdentities[0]
	}
	return nil
}

func (n *NodeEnsureRemoveAnonymousIdentity) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectRun(func(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			if n.AnonymousIdentity == nil {
				return nil
			}
			err := ctx.Identities.Delete(goCtx, n.AnonymousIdentity)
			if err != nil {
				return err
			}
			return nil
		}),
		interaction.EffectOnCommit(func(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			if n.AnonymousIdentity == nil {
				return nil
			}
			// if anonymous identity is removed during adding identity
			// it is a promotion flow
			userID := graph.MustGetUserID()
			anonUserRef := model.UserRef{
				Meta: model.Meta{
					ID: userID,
				},
			}
			var identityModels []model.Identity
			for _, info := range graph.GetUserNewIdentities() {
				identityModels = append(identityModels, info.ToModel())
			}

			err := ctx.Events.DispatchEventOnCommit(goCtx, &nonblocking.UserAnonymousPromotedEventPayload{
				AnonymousUserRef: anonUserRef,
				UserRef:          anonUserRef,
				Identities:       identityModels,
				AdminAPI:         n.IsAdminAPI,
			})
			if err != nil {
				return err
			}

			return nil
		}),
	}, nil
}

func (n *NodeEnsureRemoveAnonymousIdentity) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(goCtx, graph, n)
}
