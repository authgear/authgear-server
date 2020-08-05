package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterNode(&NodeCreateAuthenticatorBegin{})
}

type EdgeCreateAuthenticatorBegin struct {
	Stage newinteraction.AuthenticationStage
}

func (e *EdgeCreateAuthenticatorBegin) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	return &NodeCreateAuthenticatorBegin{
		Stage: e.Stage,
	}, nil
}

type NodeCreateAuthenticatorBegin struct {
	Stage newinteraction.AuthenticationStage `json:"stage"`
}

func (n *NodeCreateAuthenticatorBegin) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorBegin) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	var edges []newinteraction.Edge
	var err error

	switch n.Stage {
	case newinteraction.AuthenticationStagePrimary:
		edges, err = n.derivePrimary(ctx, graph)
		if err != nil {
			return nil, err
		}
	case newinteraction.AuthenticationStageSecondary:
		edges, err = n.deriveSecondary(ctx, graph)
		if err != nil {
			return nil, err
		}
	default:
		panic("interaction: unknown authentication stage: " + n.Stage)
	}

	// No authenticators needed, skip the stage
	if len(edges) == 0 {
		edges = append(edges, &EdgeCreateAuthenticatorEnd{Stage: n.Stage, Authenticators: nil})
	}

	return edges, nil
}

func (n *NodeCreateAuthenticatorBegin) derivePrimary(ctx *newinteraction.Context, graph *newinteraction.Graph) (edges []newinteraction.Edge, err error) {
	iden := graph.MustGetUserLastIdentity()

	// Determine whether we need to create primary authenticator.

	// 1. Check whether the identity actually requires primary authenticator.
	// If it does not, then no primary authenticator is needed.
	identityRequiresPrimaryAuthentication := len(iden.Type.PrimaryAuthenticatorTypes()) > 0
	if !identityRequiresPrimaryAuthentication {
		return nil, nil
	}

	// 2. Check what primary authenticator the developer prefers.
	// Here we check if the configuration is non-sense.
	types := ctx.Config.Authentication.PrimaryAuthenticators
	if len(types) == 0 {
		return nil, newinteraction.ConfigurationViolated.New("identity requires primary authenticator but none is enabled")
	}

	firstType := types[0]

	// 3. Find out whether the identity has the preferred primary authenticator.
	// If it does not, creation is needed.
	ais, err := ctx.Authenticators.List(
		iden.UserID,
		authenticator.KeepType(firstType),
		authenticator.KeepTag(authenticator.TagPrimaryAuthenticator),
		authenticator.KeepPrimaryAuthenticatorOfIdentity(iden),
	)
	if err != nil {
		return nil, err
	}

	if len(ais) == 0 {
		switch firstType {
		case authn.AuthenticatorTypePassword:
			edges = append(edges, &EdgeCreateAuthenticatorPassword{Stage: n.Stage})

		case authn.AuthenticatorTypeTOTP:
			edges = append(edges, &EdgeCreateAuthenticatorTOTPSetup{Stage: n.Stage})

		case authn.AuthenticatorTypeOOB:
			edges = append(edges, &EdgeCreateAuthenticatorOOBSetup{Stage: n.Stage})
		default:
			panic("interaction: unknown authenticator type: " + firstType)
		}
	}

	return edges, nil
}

func (n *NodeCreateAuthenticatorBegin) deriveSecondary(ctx *newinteraction.Context, graph *newinteraction.Graph) (edges []newinteraction.Edge, err error) {
	var requiredType authn.AuthenticatorType

	userID := graph.MustGetUserID()

	ais, err := ctx.Authenticators.List(
		userID,
		authenticator.KeepTag(authenticator.TagSecondaryAuthenticator),
	)

	mode := ctx.Config.Authentication.SecondaryAuthenticationMode
	types := ctx.Config.Authentication.SecondaryAuthenticators

	// FIXME(mfa): Right now we only consider signup
	if mode == config.SecondaryAuthenticationModeRequired && len(types) > 0 {
		first := types[0]

		found := false
		for _, ai := range ais {
			if ai.Type == first {
				found = true
				break
			}
		}

		if !found {
			requiredType = first
		}
	}

	if requiredType != "" {
		switch requiredType {
		case authn.AuthenticatorTypePassword:
			edges = append(edges, &EdgeCreateAuthenticatorPassword{Stage: n.Stage})

		case authn.AuthenticatorTypeTOTP:
			edges = append(edges, &EdgeCreateAuthenticatorTOTPSetup{Stage: n.Stage})

		case authn.AuthenticatorTypeOOB:
			edges = append(edges, &EdgeCreateAuthenticatorOOBSetup{Stage: n.Stage})
		default:
			panic("interaction: unknown authenticator type: " + requiredType)
		}
	}

	return edges, nil
}
