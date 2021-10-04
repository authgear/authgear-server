package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeDoCreateIdentity{})
}

type EdgeDoCreateIdentity struct {
	Identity   *identity.Info
	IsAddition bool
}

func (e *EdgeDoCreateIdentity) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	isAdminAPI := interaction.IsAdminAPI(rawInput)
	modifyDisabled := e.Identity.ModifyDisabled(ctx.Config.Identity)
	if e.IsAddition && !isAdminAPI && modifyDisabled {
		return nil, interaction.ErrIdentityModifyDisabled
	}
	return &NodeDoCreateIdentity{
		Identity:   e.Identity,
		IsAddition: e.IsAddition,
		IsAdminAPI: isAdminAPI,
	}, nil
}

type NodeDoCreateIdentity struct {
	Identity   *identity.Info `json:"identity"`
	IsAddition bool           `json:"is_addition"`
	IsAdminAPI bool           `json:"is_admin_api"`
}

func (n *NodeDoCreateIdentity) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoCreateIdentity) GetEffects() ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectRun(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			user, err := ctx.Users.Get(n.Identity.UserID)
			if err != nil {
				return err
			}

			if n.Identity.Type == authn.IdentityTypeBiometric && user.IsAnonymous {
				return interaction.NewInvariantViolated(
					"AnonymousUserAddIdentity",
					"anonymous user cannot add identity",
					nil,
				)
			}

			if _, err := ctx.Identities.CheckDuplicated(n.Identity); err != nil {
				if errors.Is(err, identity.ErrIdentityAlreadyExists) {
					return n.Identity.FillDetails(interaction.ErrDuplicatedIdentity)
				}
				return err
			}
			if err := ctx.Identities.Create(n.Identity); err != nil {
				return err
			}

			if !n.IsAddition && ctx.Config.UserProfile.StandardAttributes.Population.Strategy == config.StandardAttributesPopulationStrategyOnSignup {
				err := ctx.Users.PopulateStandardAttributes(n.Identity.UserID, n.Identity)
				if err != nil {
					return err
				}
			}

			return nil
		}),
		interaction.EffectOnCommit(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			if _, creating := graph.GetNewUserID(); creating {
				return nil
			}

			user, err := ctx.Users.Get(n.Identity.UserID)
			if err != nil {
				return err
			}

			var e event.Payload
			switch n.Identity.Type {
			case authn.IdentityTypeLoginID:
				loginIDType := n.Identity.Claims[identity.IdentityClaimLoginIDType].(string)
				e = nonblocking.NewIdentityLoginIDAddedEventPayload(
					*user,
					n.Identity.ToModel(),
					loginIDType,
					n.IsAdminAPI,
				)
			case authn.IdentityTypeOAuth:
				e = &nonblocking.IdentityOAuthConnectedEventPayload{
					User:     *user,
					Identity: n.Identity.ToModel(),
					AdminAPI: n.IsAdminAPI,
				}
			}

			if e != nil {
				err = ctx.Events.DispatchEvent(e)
				if err != nil {
					return err
				}
			}

			return nil
		}),
		interaction.EffectOnCommit(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			if _, creating := graph.GetNewUserID(); creating {
				return nil
			}

			err := ctx.Search.ReindexUser(n.Identity.UserID, false)
			if err != nil {
				return err
			}
			return nil
		}),
	}, nil
}

func (n *NodeDoCreateIdentity) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}

func (n *NodeDoCreateIdentity) UserNewIdentity() *identity.Info {
	return n.Identity
}
