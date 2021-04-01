package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeDoCreateIdentity{})
}

type EdgeDoCreateIdentity struct {
	Identity *identity.Info
}

func (e *EdgeDoCreateIdentity) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	return &NodeDoCreateIdentity{
		Identity:   e.Identity,
		IsAdminAPI: interaction.IsAdminAPI(rawInput),
	}, nil
}

type NodeDoCreateIdentity struct {
	Identity   *identity.Info `json:"identity"`
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
					return interaction.ErrDuplicatedIdentity
				}
				return err
			}
			if err := ctx.Identities.Create(n.Identity); err != nil {
				return err
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
				e = nonblocking.NewIdentityLoginIDAddedEvent(
					*user,
					n.Identity.ToModel(),
					loginIDType,
				)
			case authn.IdentityTypeOAuth:
				e = &nonblocking.IdentityOAuthConnectedEvent{
					User:     *user,
					Identity: n.Identity.ToModel(),
				}
			}

			if e != nil {
				err = ctx.Hooks.DispatchEvent(e)
				if err != nil {
					return err
				}
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
