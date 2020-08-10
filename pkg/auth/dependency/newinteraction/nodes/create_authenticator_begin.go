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
	Stage                newinteraction.AuthenticationStage `json:"stage"`
	AuthenticationConfig *config.AuthenticationConfig       `json:"-"`
	Authenticators       []*authenticator.Info              `json:"-"`
}

func (n *NodeCreateAuthenticatorBegin) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	ais, err := ctx.Authenticators.List(graph.MustGetUserID())
	if err != nil {
		return err
	}

	n.AuthenticationConfig = ctx.Config.Authentication
	n.Authenticators = ais
	return nil
}

func (n *NodeCreateAuthenticatorBegin) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorBegin) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	var edges []newinteraction.Edge
	var err error

	switch n.Stage {
	case newinteraction.AuthenticationStagePrimary:
		edges, err = n.derivePrimary(graph)
		if err != nil {
			return nil, err
		}
	case newinteraction.AuthenticationStageSecondary:
		edges, err = n.deriveSecondary(graph)
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

func (n *NodeCreateAuthenticatorBegin) derivePrimary(graph *newinteraction.Graph) (edges []newinteraction.Edge, err error) {
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
	types := n.AuthenticationConfig.PrimaryAuthenticators
	if len(types) == 0 {
		return nil, newinteraction.ConfigurationViolated.New("identity requires primary authenticator but none is enabled")
	}

	firstType := types[0]

	// 3. Find out whether the identity has the preferred primary authenticator.
	// If it does not, creation is needed.
	ais := filterAuthenticators(
		n.Authenticators,
		authenticator.KeepType(firstType),
		authenticator.KeepTag(authenticator.TagPrimaryAuthenticator),
		authenticator.KeepPrimaryAuthenticatorOfIdentity(iden),
	)

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

func (n *NodeCreateAuthenticatorBegin) deriveSecondary(graph *newinteraction.Graph) (edges []newinteraction.Edge, err error) {

	// Determine whether we need to create secondary authenticator.

	// 1. Check secondary authentication mode.
	// If it is not required, then no secondary authenticator is needed.
	// FIXME(mfa): Right now we only consider signup/login.
	mode := n.AuthenticationConfig.SecondaryAuthenticationMode
	if mode != config.SecondaryAuthenticationModeRequired {
		return nil, nil
	}

	// 2. Check whether
	//   the set of secondary authenticators of the user, and
	//   the set of preferred secondary authenticators
	// have intersection.
	// If there is no intersection, create the first preferred one.
	// Here we also check for non-sense configuration
	types := n.AuthenticationConfig.SecondaryAuthenticators
	if len(types) == 0 {
		return nil, newinteraction.ConfigurationViolated.New("secondary authentication is required but no secondary authenticator is enabled")
	}

	ais := filterAuthenticators(
		n.Authenticators,
		authenticator.KeepTag(authenticator.TagSecondaryAuthenticator),
	)
	if err != nil {
		return nil, err
	}

	intersection := make(map[authn.AuthenticatorType]struct{})
	for _, typ := range types {
		for _, a := range ais {
			if a.Type == typ {
				intersection[typ] = struct{}{}
			}
		}
	}

	// FIXME(mfa): Allow the user to choose between which secondary authenticator to setup.
	// Right now, EdgeCreateAuthenticatorTOTPSetup always instantiate without any input.
	if len(intersection) == 0 {
		firstType := types[0]
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
