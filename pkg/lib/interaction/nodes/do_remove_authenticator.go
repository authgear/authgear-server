package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeDoRemoveAuthenticator{})
}

type EdgeDoRemoveAuthenticator struct {
	Authenticator        *authenticator.Info
	BypassMFARequirement bool
}

func (e *EdgeDoRemoveAuthenticator) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	isAdminAPI := interaction.IsAdminAPI(rawInput)
	return &NodeDoRemoveAuthenticator{
		Authenticator:        e.Authenticator,
		BypassMFARequirement: e.BypassMFARequirement,
		IsAdminAPI:           isAdminAPI,
	}, nil
}

type NodeDoRemoveAuthenticator struct {
	Authenticator        *authenticator.Info `json:"authenticator"`
	BypassMFARequirement bool                `json:"bypass_mfa_requirement"`
	IsAdminAPI           bool                `json:"is_admin_api"`
}

func (n *NodeDoRemoveAuthenticator) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

// nolint:gocognit
func (n *NodeDoRemoveAuthenticator) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectRun(func(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			userID := graph.MustGetUserID()

			as, err := ctx.Authenticators.List(goCtx, userID)
			if err != nil {
				return err
			}

			switch n.Authenticator.Kind {
			case authenticator.KindPrimary:
				if n.Authenticator.Type == model.AuthenticatorTypePasskey {
					return api.NewInvariantViolated(
						"RemovePasskeyAuthenticator",
						"cannot delete passkey authenticator, should delete passkey identity instead",
						nil,
					)
				}

				// Ensure all identities have matching primary authenticator.
				is, err := ctx.Identities.ListByUser(goCtx, userID)
				if err != nil {
					return err
				}

				// Admin is allowed to remove the last authenticator
				if n.IsAdminAPI {
					break
				}
				for _, i := range is {
					primaryAuths := authenticator.ApplyFilters(as, authenticator.KeepPrimaryAuthenticatorOfIdentity(i))
					if len(primaryAuths) == 1 && primaryAuths[0].ID == n.Authenticator.ID {
						return api.NewInvariantViolated(
							"RemoveLastPrimaryAuthenticator",
							"cannot remove last primary authenticator for identity",
							map[string]interface{}{"identity_id": i.ID},
						)
					}
				}

			case authenticator.KindSecondary:
				// Ensure authenticators conform to MFA requirement configuration
				if n.BypassMFARequirement {
					break
				}
				// Admin is allowed to remove the last authenticator
				if n.IsAdminAPI {
					break
				}
				primaries := authenticator.ApplyFilters(as, authenticator.KeepPrimaryAuthenticatorCanHaveMFA)
				secondaries := authenticator.ApplyFilters(as, authenticator.KeepKind(authenticator.KindSecondary))
				mode := ctx.Config.Authentication.SecondaryAuthenticationMode

				cannotRemove := mode == config.SecondaryAuthenticationModeRequired &&
					len(primaries) > 0 &&
					len(secondaries) == 1 && secondaries[0].ID == n.Authenticator.ID

				if cannotRemove {
					return api.NewInvariantViolated(
						"RemoveLastSecondaryAuthenticator",
						"cannot remove last secondary authenticator",
						nil,
					)
				}
			}

			err = ctx.Authenticators.Delete(goCtx, n.Authenticator)
			if err != nil {
				return err
			}

			return nil
		}),
	}, nil
}

func (n *NodeDoRemoveAuthenticator) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(goCtx, graph, n)
}
