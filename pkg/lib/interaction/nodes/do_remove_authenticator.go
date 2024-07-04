package nodes

import (
	"fmt"

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

func (e *EdgeDoRemoveAuthenticator) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	return &NodeDoRemoveAuthenticator{
		Authenticator:        e.Authenticator,
		BypassMFARequirement: e.BypassMFARequirement,
	}, nil
}

type NodeDoRemoveAuthenticator struct {
	Authenticator        *authenticator.Info `json:"authenticator"`
	BypassMFARequirement bool                `json:"bypass_mfa_requirement"`
}

func (n *NodeDoRemoveAuthenticator) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

// nolint:gocognit
func (n *NodeDoRemoveAuthenticator) GetEffects() ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectRun(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			userID := graph.MustGetUserID()

			as, err := ctx.Authenticators.List(userID)
			if err != nil {
				return err
			}

			switch n.Authenticator.Kind {
			case authenticator.KindPrimary:
				if n.Authenticator.Type == model.AuthenticatorTypePasskey {
					return fmt.Errorf("cannot delete passkey authenticator, should delete passkey identity instead")
				}

				// Ensure all identities have matching primary authenticator.
				is, err := ctx.Identities.ListByUser(userID)
				if err != nil {
					return err
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

			err = ctx.Authenticators.Delete(n.Authenticator)
			if err != nil {
				return err
			}

			return nil
		}),
	}, nil
}

func (n *NodeDoRemoveAuthenticator) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
