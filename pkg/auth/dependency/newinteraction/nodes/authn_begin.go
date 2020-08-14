package nodes

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	newinteraction.RegisterNode(&NodeAuthenticationBegin{})
}

type EdgeAuthenticationBegin struct {
	Stage newinteraction.AuthenticationStage
}

func (e *EdgeAuthenticationBegin) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	return &NodeAuthenticationBegin{
		Stage: e.Stage,
	}, nil
}

type NodeAuthenticationBegin struct {
	Stage                newinteraction.AuthenticationStage `json:"stage"`
	Identity             *identity.Info                     `json:"-"`
	AuthenticationConfig *config.AuthenticationConfig       `json:"-"`
	Authenticators       []*authenticator.Info              `json:"-"`
}

func (n *NodeAuthenticationBegin) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	ais, err := ctx.Authenticators.List(graph.MustGetUserID())
	if err != nil {
		return err
	}

	n.Identity = graph.MustGetUserLastIdentity()
	n.AuthenticationConfig = ctx.Config.Authentication
	n.Authenticators = ais
	return nil
}

func (n *NodeAuthenticationBegin) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeAuthenticationBegin) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return n.GetAuthenticationEdges(), nil
}

// GetAuthenticationEdges implements AuthenticationBeginNode.
func (n *NodeAuthenticationBegin) GetAuthenticationEdges() []newinteraction.Edge {
	var edges []newinteraction.Edge
	var availableAuthenticators []*authenticator.Info
	var preferred []authn.AuthenticatorType

	switch n.Stage {
	case newinteraction.AuthenticationStagePrimary:
		preferred = n.AuthenticationConfig.PrimaryAuthenticators
		availableAuthenticators = filterAuthenticators(
			n.Authenticators,
			authenticator.KeepTag(authenticator.TagPrimaryAuthenticator),
			authenticator.KeepPrimaryAuthenticatorOfIdentity(n.Identity),
		)
	case newinteraction.AuthenticationStageSecondary:
		preferred = n.AuthenticationConfig.SecondaryAuthenticators
		availableAuthenticators = filterAuthenticators(
			n.Authenticators,
			authenticator.KeepTag(authenticator.TagSecondaryAuthenticator),
		)
	default:
		panic("interaction: unknown authentication stage: " + n.Stage)
	}

	passwords := filterAuthenticators(
		availableAuthenticators,
		authenticator.KeepType(authn.AuthenticatorTypePassword),
	)
	newinteraction.SortAuthenticators(
		nil,
		passwords,
		func(i int) newinteraction.SortableAuthenticator {
			a := newinteraction.SortableAuthenticatorInfo(*passwords[i])
			return &a
		},
	)

	totps := filterAuthenticators(
		availableAuthenticators,
		authenticator.KeepType(authn.AuthenticatorTypeTOTP),
	)
	newinteraction.SortAuthenticators(
		nil,
		totps,
		func(i int) newinteraction.SortableAuthenticator {
			a := newinteraction.SortableAuthenticatorInfo(*totps[i])
			return &a
		},
	)

	oobs := filterAuthenticators(
		availableAuthenticators,
		authenticator.KeepType(authn.AuthenticatorTypeOOB),
	)
	newinteraction.SortAuthenticators(
		nil,
		totps,
		func(i int) newinteraction.SortableAuthenticator {
			a := newinteraction.SortableAuthenticatorInfo(*oobs[i])
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

	// No authenticators found, skip the authentication stage
	if len(edges) == 0 {
		edges = append(edges, &EdgeAuthenticationEnd{
			Stage:  n.Stage,
			Result: AuthenticationResultOptional,
		})
		return edges
	}

	newinteraction.SortAuthenticators(
		preferred,
		edges,
		func(i int) newinteraction.SortableAuthenticator {
			edge := edges[i]
			a, ok := edge.(newinteraction.SortableAuthenticator)
			if !ok {
				panic(fmt.Sprintf("interaction: unknown edge: %T", edge))
			}
			return a
		},
	)

	if n.Stage == newinteraction.AuthenticationStageSecondary {
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

	return edges
}
