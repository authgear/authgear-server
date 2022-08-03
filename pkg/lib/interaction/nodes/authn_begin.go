package nodes

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
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
	Stage authn.AuthenticationStage
}

func (e *EdgeAuthenticationBegin) Instantiate(ctx *interaction.Context, graph *interaction.Graph, input interface{}) (interaction.Node, error) {
	return &NodeAuthenticationBegin{
		Stage: e.Stage,
	}, nil
}

type NodeAuthenticationBegin struct {
	Stage                authn.AuthenticationStage    `json:"stage"`
	Identity             *identity.Info               `json:"-"`
	PrimaryAuthenticator *authenticator.Info          `json:"-"`
	AuthenticationConfig *config.AuthenticationConfig `json:"-"`
	AuthenticatorConfig  *config.AuthenticatorConfig  `json:"-"`
	Authenticators       []*authenticator.Info        `json:"-"`
}

func (n *NodeAuthenticationBegin) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	ais, err := ctx.Authenticators.List(graph.MustGetUserID())
	if err != nil {
		return err
	}

	n.Identity = graph.MustGetUserLastIdentity()
	if prim, ok := graph.GetUserAuthenticator(authn.AuthenticationStagePrimary); ok {
		n.PrimaryAuthenticator = prim
	}
	n.AuthenticationConfig = ctx.Config.Authentication
	n.AuthenticatorConfig = ctx.Config.Authenticator
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
func (n *NodeAuthenticationBegin) GetAuthenticationStage() authn.AuthenticationStage {
	return n.Stage
}

// GetAuthenticationEdges implements AuthenticationBeginNode.
func (n *NodeAuthenticationBegin) GetAuthenticationEdges() ([]interaction.Edge, error) {
	var edges []interaction.Edge
	var availableAuthenticators []*authenticator.Info
	var preferred []model.AuthenticatorType
	var required []model.AuthenticatorType

	switch n.Stage {
	case authn.AuthenticationStagePrimary:
		preferred = *n.AuthenticationConfig.PrimaryAuthenticators
		availableAuthenticators = authenticator.ApplyFilters(
			n.Authenticators,
			authenticator.KeepPrimaryAuthenticatorOfIdentity(n.Identity),
		)
		required = n.Identity.PrimaryAuthenticatorTypes()

	case authn.AuthenticationStageSecondary:
		preferred = *n.AuthenticationConfig.SecondaryAuthenticators
		if n.PrimaryAuthenticator.CanHaveMFA() {
			availableAuthenticators = authenticator.ApplyFilters(
				n.Authenticators,
				authenticator.KeepKind(authenticator.KindSecondary),
			)
		}

		switch n.AuthenticationConfig.SecondaryAuthenticationMode {
		case config.SecondaryAuthenticationModeIfExists,
			config.SecondaryAuthenticationModeRequired:
			// Require secondary authentication if any authenticator exists.
			// For Required mode: treat same as IfExists, so user will not
			// be locked out when requirement changed.
			existingAuths := map[model.AuthenticatorType]struct{}{}
			for _, a := range availableAuthenticators {
				existingAuths[a.Type] = struct{}{}
			}
			for t := range existingAuths {
				required = append(required, t)
			}
		case config.SecondaryAuthenticationModeDisabled:
			// When MFA is disabled, treat it as if the user has no authenticators.
			availableAuthenticators = nil
			required = nil
		}

	default:
		panic("interaction: unknown authentication stage: " + n.Stage)
	}

	passwords := authenticator.ApplyFilters(
		availableAuthenticators,
		authenticator.KeepType(model.AuthenticatorTypePassword),
	)
	interaction.SortAuthenticators(
		nil,
		passwords,
		func(i int) interaction.SortableAuthenticator {
			a := interaction.SortableAuthenticatorInfo(*passwords[i])
			return &a
		},
	)

	totps := authenticator.ApplyFilters(
		availableAuthenticators,
		authenticator.KeepType(model.AuthenticatorTypeTOTP),
	)
	interaction.SortAuthenticators(
		nil,
		totps,
		func(i int) interaction.SortableAuthenticator {
			a := interaction.SortableAuthenticatorInfo(*totps[i])
			return &a
		},
	)

	emailoobs := authenticator.ApplyFilters(
		availableAuthenticators,
		authenticator.KeepType(model.AuthenticatorTypeOOBEmail),
	)
	interaction.SortAuthenticators(
		nil,
		emailoobs,
		func(i int) interaction.SortableAuthenticator {
			a := interaction.SortableAuthenticatorInfo(*emailoobs[i])
			return &a
		},
	)

	smsoobs := authenticator.ApplyFilters(
		availableAuthenticators,
		authenticator.KeepType(model.AuthenticatorTypeOOBSMS),
	)
	interaction.SortAuthenticators(
		nil,
		smsoobs,
		func(i int) interaction.SortableAuthenticator {
			a := interaction.SortableAuthenticatorInfo(*smsoobs[i])
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

	if len(emailoobs) > 0 {
		edges = append(edges, &EdgeAuthenticationOOBTrigger{
			Stage:                n.Stage,
			Authenticators:       emailoobs,
			OOBAuthenticatorType: model.AuthenticatorTypeOOBEmail,
		})
	}

	if len(smsoobs) > 0 {
		if n.AuthenticatorConfig.OOB.SMS.PhoneOTPMode.IsWhatsappEnabled() {
			edges = append(edges, &EdgeAuthenticationWhatsappTrigger{
				Stage:          n.Stage,
				Authenticators: smsoobs,
			})
		}

		if n.AuthenticatorConfig.OOB.SMS.PhoneOTPMode.IsSMSEnabled() {
			edges = append(edges, &EdgeAuthenticationOOBTrigger{
				Stage:                n.Stage,
				Authenticators:       smsoobs,
				OOBAuthenticatorType: model.AuthenticatorTypeOOBSMS,
			})
		}
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
				Stage:              n.Stage,
				AuthenticationType: authn.AuthenticationTypeNone,
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

	if n.Stage == authn.AuthenticationStageSecondary {
		// If we reach here, there are at least one secondary authenticator.
		// We allow the use of recovery code if it is not disabled.
		// We have to add after the sorting because
		// recovery code is not an authenticator.
		if !*n.AuthenticationConfig.RecoveryCode.Disabled {
			edges = append(edges, &EdgeConsumeRecoveryCode{})
		}

		// Allow the use of device token.
		if !n.AuthenticationConfig.DeviceToken.Disabled {
			edges = append(edges, &EdgeUseDeviceToken{})
		}
	}

	return edges, nil
}
