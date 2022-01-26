package nodes

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeEnsureRemoveAnonymousIdentity{})
}

type EdgeEnsureRemoveAnonymousIdentity struct{}

func (e *EdgeEnsureRemoveAnonymousIdentity) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	return &NodeEnsureRemoveAnonymousIdentity{}, nil
}

type NodeEnsureRemoveAnonymousIdentity struct {
	AnonymousIdentity *identity.Info `json:"anonymous_identity,omitempty"`
}

func (n *NodeEnsureRemoveAnonymousIdentity) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	userID := graph.MustGetUserID()
	iis, err := ctx.Identities.ListByUser(graph.MustGetUserID())
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
		n.AnonymousIdentity = anonymousIdentities[0]
	}
	return nil
}

func (n *NodeEnsureRemoveAnonymousIdentity) GetEffects() ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectRun(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			if n.AnonymousIdentity == nil {
				return nil
			}
			userID := graph.MustGetUserID()
			identities, err := ctx.Identities.ListByUser(userID)
			if err != nil {
				return err
			}
			remaining := identity.ApplyFilters(
				identities,
				identity.KeepIdentifiable,
				identity.OmitID(n.AnonymousIdentity.ID),
			)
			if len(remaining) < 1 {
				// This node should be run after adding a identifiable identity
				panic("interaction: missing identifiable identities")
			}
			err = ctx.Identities.Delete(n.AnonymousIdentity)
			if err != nil {
				return err
			}
			return nil
		}),
	}, nil
}

func (n *NodeEnsureRemoveAnonymousIdentity) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
