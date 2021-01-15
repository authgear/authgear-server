package nodes

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeAuthenticationBegin{})
}

type EdgeAuthenticationBegin struct {
	Stage interaction.AuthenticationStage
}

func (e *EdgeAuthenticationBegin) Instantiate(ctx *interaction.Context, graph *interaction.Graph, input interface{}) (interaction.Node, error) {
	return &NodeAuthenticationBegin{
		Stage: e.Stage,
	}, nil
}

type NodeAuthenticationBegin struct {
	Stage                interaction.AuthenticationStage `json:"stage"`
	Identity             *identity.Info                  `json:"-"`
	AuthenticationConfig *config.AuthenticationConfig    `json:"-"`
	Authenticators       []*authenticator.Info           `json:"-"`
}

func (n *NodeAuthenticationBegin) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	ais, err := ctx.Authenticators.List(graph.MustGetUserID())
	if err != nil {
		return err
	}

	n.Identity = graph.MustGetUserLastIdentity()
	n.AuthenticationConfig = ctx.Config.Authentication
	n.Authenticators = ais
	return nil
}

func (n *NodeAuthenticationBegin) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeAuthenticationBegin) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return n.GetAuthenticationEdges()
}

// GetAuthenticationStage implements AuthenticationBeginNode.
func (n *NodeAuthenticationBegin) GetAuthenticationStage() interaction.AuthenticationStage {
	return n.Stage
}

// GetAuthenticationEdges implements AuthenticationBeginNode.
func (n *NodeAuthenticationBegin) GetAuthenticationEdges() ([]interaction.Edge, error) {
	var edges []interaction.Edge
	var availableAuthenticators []*authenticator.Info
	var preferred []authn.AuthenticatorType
	var required []authn.AuthenticatorType

	switch n.Stage {
	case interaction.AuthenticationStagePrimary:
		preferred = n.AuthenticationConfig.PrimaryAuthenticators
		availableAuthenticators = filterAuthenticators(
			n.Authenticators,
			authenticator.KeepPrimaryAuthenticatorOfIdentity(n.Identity),
		)
		required = n.Identity.PrimaryAuthenticatorTypes()

	case interaction.AuthenticationStageSecondary:
		preferred = n.AuthenticationConfig.SecondaryAuthenticators
		availableAuthenticators = filterAuthenticators(
			n.Authenticators,
			authenticator.KeepKind(authenticator.KindSecondary),
		)

		switch n.AuthenticationConfig.SecondaryAuthenticationMode {
		case config.SecondaryAuthenticationModeIfExists,
			config.SecondaryAuthenticationModeRequired:
			// Require secondary authentication if any authenticator exists.
			// For Required mode: treat same as IfExists, so user will not
			// be locked out when requirement changed.
			existingAuths := map[authn.AuthenticatorType]struct{}{}
			for _, a := range availableAuthenticators {
				existingAuths[a.Type] = struct{}{}
			}
			for t := range existingAuths {
				required = append(required, t)
			}

		case config.SecondaryAuthenticationModeIfRequested:
			// Never require unless explicitly requested.
			required = nil
		}

	default:
		panic("interaction: unknown authentication stage: " + n.Stage)
	}

	passwords := filterAuthenticators(
		availableAuthenticators,
		authenticator.KeepType(authn.AuthenticatorTypePassword),
	)
	interaction.SortAuthenticators(
		nil,
		passwords,
		func(i int) interaction.SortableAuthenticator {
			a := interaction.SortableAuthenticatorInfo(*passwords[i])
			return &a
		},
	)

	totps := filterAuthenticators(
		availableAuthenticators,
		authenticator.KeepType(authn.AuthenticatorTypeTOTP),
	)
	interaction.SortAuthenticators(
		nil,
		totps,
		func(i int) interaction.SortableAuthenticator {
			a := interaction.SortableAuthenticatorInfo(*totps[i])
			return &a
		},
	)

	oobs := filterAuthenticators(
		availableAuthenticators,
		authenticator.KeepType(authn.AuthenticatorTypeOOB),
	)
	interaction.SortAuthenticators(
		nil,
		totps,
		func(i int) interaction.SortableAuthenticator {
			a := interaction.SortableAuthenticatorInfo(*oobs[i])
			return &a
		},
	)

	if len(passwords) > 0 {
		edges = append(edges, &EdgeAuthenticationPassword{
			Stage:          n.Stage,
			Authenticators: passwords,
		})
	}

	if len(totps) > 0 {
		edges = append(edges, &EdgeAuthenticationTOTP{
			Stage:          n.Stage,
			Authenticators: totps,
		})
	}

	if len(oobs) > 0 {
		edges = append(edges, &EdgeAuthenticationOOBTrigger{
			Stage:          n.Stage,
			Authenticators: oobs,
		})
	}

	// No authenticators found, skip the authentication stage if not required.
	// If identity requires authentication, the identity cannot be authenticated.
	if len(edges) == 0 {
		if len(required) > 0 {
			return nil, interaction.NewInvariantViolated(
				"MissingAuthenticator",
				"missing authenticator for the identity",
				nil,
			)
		}

		return []interaction.Edge{
			&EdgeAuthenticationEnd{
				Stage:  n.Stage,
				Result: AuthenticationResultOptional,
			},
		}, nil
	}

	interaction.SortAuthenticators(
		preferred,
		edges,
		func(i int) interaction.SortableAuthenticator {
			edge := edges[i]
			a, ok := edge.(interaction.SortableAuthenticator)
			if !ok {
				panic(fmt.Sprintf("interaction: unknown edge: %T", edge))
			}
			return a
		},
	)

	if n.Stage == interaction.AuthenticationStageSecondary {
		// If we reach here, there are at least one secondary authenticator
		// so we have to allow the use of recovery code.
		// We have to add after the sorting because
		// recovery code is not an authenticator.
		edges = append(edges, &EdgeConsumeRecoveryCode{})

		// Allow the use of device token.
		if !n.AuthenticationConfig.DeviceToken.Disabled {
			edges = append(edges, &EdgeUseDeviceToken{})
		}
	}

	return edges, nil
}
