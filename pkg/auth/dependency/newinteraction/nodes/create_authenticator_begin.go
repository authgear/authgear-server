package nodes

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
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
	Identity             *identity.Info                     `json:"-"`
	AuthenticationConfig *config.AuthenticationConfig       `json:"-"`
	AuthenticatorConfig  *config.AuthenticatorConfig        `json:"-"`
	Authenticators       []*authenticator.Info              `json:"-"`
}

func (n *NodeCreateAuthenticatorBegin) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	ais, err := ctx.Authenticators.List(graph.MustGetUserID())
	if err != nil {
		return err
	}

	n.Identity = graph.MustGetUserLastIdentity()
	n.AuthenticationConfig = ctx.Config.Authentication
	n.AuthenticatorConfig = ctx.Config.Authenticator
	n.Authenticators = ais
	return nil
}

func (n *NodeCreateAuthenticatorBegin) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorBegin) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return n.deriveEdges()
}

// GetAllowedChannels implements SetupOOBOTPNode.
func (n *NodeCreateAuthenticatorBegin) GetAllowedChannels() ([]authn.AuthenticatorOOBChannel, error) {
	edges, err := n.deriveEdges()
	if err != nil {
		return nil, err
	}

	for _, edge := range edges {
		switch edge := edge.(type) {
		case *EdgeCreateAuthenticatorOOBSetup:
			return edge.AllowedChannels, nil
		}
	}

	return nil, fmt.Errorf("interaction: expected to find EdgeCreateAuthenticatorOOBSetup")
}

// GetCreateAuthenticatorEdges implements CreateAuthenticatorBeginNode.
func (n *NodeCreateAuthenticatorBegin) GetCreateAuthenticatorEdges() ([]newinteraction.Edge, error) {
	return n.deriveEdges()
}

func (n *NodeCreateAuthenticatorBegin) deriveEdges() ([]newinteraction.Edge, error) {
	var edges []newinteraction.Edge
	var err error

	switch n.Stage {
	case newinteraction.AuthenticationStagePrimary:
		edges, err = n.derivePrimary()
		if err != nil {
			return nil, err
		}
	case newinteraction.AuthenticationStageSecondary:
		edges = n.deriveSecondary()
	default:
		panic("interaction: unknown authentication stage: " + n.Stage)
	}

	// No authenticators needed, skip the stage
	if len(edges) == 0 {
		edges = append(edges, &EdgeCreateAuthenticatorEnd{Stage: n.Stage, Authenticators: nil})
	}

	return edges, nil
}

func (n *NodeCreateAuthenticatorBegin) derivePrimary() (edges []newinteraction.Edge, err error) {
	// Determine whether we need to create primary authenticator.

	// 1. Check whether the identity actually requires primary authenticator.
	// If it does not, then no primary authenticator is needed.
	identityRequiresPrimaryAuthentication := len(n.Identity.Type.PrimaryAuthenticatorTypes()) > 0
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
		authenticator.KeepPrimaryAuthenticatorOfIdentity(n.Identity),
	)

	if len(ais) == 0 {
		switch firstType {
		case authn.AuthenticatorTypePassword:
			edges = append(edges, &EdgeCreateAuthenticatorPassword{Stage: n.Stage})

		case authn.AuthenticatorTypeTOTP:
			edges = append(edges, &EdgeCreateAuthenticatorTOTPSetup{Stage: n.Stage})

		case authn.AuthenticatorTypeOOB:
			edges = append(edges, &EdgeCreateAuthenticatorOOBSetup{
				Stage: n.Stage,
				// To make things simpler, just include all channels.
				AllowedChannels: []authn.AuthenticatorOOBChannel{
					authn.AuthenticatorOOBChannelEmail,
					authn.AuthenticatorOOBChannelSMS,
				},
			})
		default:
			panic("interaction: unknown authenticator type: " + firstType)
		}
	}

	return edges, nil
}

func (n *NodeCreateAuthenticatorBegin) deriveSecondary() (edges []newinteraction.Edge) {
	// Determine whether we need to create secondary authenticator.

	// 1. Check secondary authentication mode.
	// If it is not required, then no secondary authenticator is needed.
	mode := n.AuthenticationConfig.SecondaryAuthenticationMode
	if mode != config.SecondaryAuthenticationModeRequired {
		return nil
	}

	// 2. Determine what secondary authenticator we allow the user to create.
	// We have the following conditions to hold:
	//   A. The secondary authenticator is allowed in the configuration.
	//   B. The user does not have that type of secondary authenticator yet.
	//   C. The number of the secondary authenticator the user is less than maximum.

	ais := filterAuthenticators(
		n.Authenticators,
		authenticator.KeepTag(authenticator.TagSecondaryAuthenticator),
	)

	passwordCount := 0
	totpCount := 0
	oobSMSCount := 0
	oobEmailCount := 0
	for _, a := range ais {
		switch a.Type {
		case authn.AuthenticatorTypePassword:
			passwordCount++
		case authn.AuthenticatorTypeTOTP:
			totpCount++
		case authn.AuthenticatorTypeOOB:
			channel := a.Props[authenticator.AuthenticatorPropOOBOTPChannelType].(string)
			switch authn.AuthenticatorOOBChannel(channel) {
			case authn.AuthenticatorOOBChannelEmail:
				oobEmailCount++
			case authn.AuthenticatorOOBChannelSMS:
				oobSMSCount++
			default:
				panic("interaction: unknown OOB channel: " + channel)
			}
		default:
			panic("interaction: unknown authenticator type: " + a.Type)
		}
	}

	// Condition A.
	for _, typ := range n.AuthenticationConfig.SecondaryAuthenticators {
		switch typ {
		case authn.AuthenticatorTypePassword:
			count := passwordCount
			// Condition B.
			if count == 0 {
				edges = append(edges, &EdgeCreateAuthenticatorPassword{Stage: n.Stage})
			}
		case authn.AuthenticatorTypeTOTP:
			// Condition B and C.
			if totpCount == 0 && totpCount < *n.AuthenticatorConfig.TOTP.Maximum {
				edges = append(edges, &EdgeCreateAuthenticatorTOTPSetup{Stage: n.Stage})
			}
		case authn.AuthenticatorTypeOOB:
			var allowedChannels []authn.AuthenticatorOOBChannel
			// Condition B and C.
			if oobSMSCount == 0 && oobSMSCount < *n.AuthenticatorConfig.OOB.SMS.Maximum {
				allowedChannels = append(allowedChannels, authn.AuthenticatorOOBChannelSMS)
			}
			// Condition B and C.
			if oobEmailCount == 0 && oobEmailCount < *n.AuthenticatorConfig.OOB.Email.Maximum {
				allowedChannels = append(allowedChannels, authn.AuthenticatorOOBChannelEmail)
			}
			if len(allowedChannels) > 0 {
				edges = append(edges, &EdgeCreateAuthenticatorOOBSetup{
					Stage:           n.Stage,
					AllowedChannels: allowedChannels,
				})
			}
		default:
			panic("interaction: unknown authenticator type: " + typ)
		}
	}

	newinteraction.SortAuthenticators(
		n.AuthenticationConfig.SecondaryAuthenticators,
		edges,
		func(i int) authn.AuthenticatorType {
			edge := edges[i]
			switch edge.(type) {
			case *EdgeCreateAuthenticatorPassword:
				return authn.AuthenticatorTypePassword
			case *EdgeCreateAuthenticatorTOTPSetup:
				return authn.AuthenticatorTypeTOTP
			case *EdgeCreateAuthenticatorOOBSetup:
				return authn.AuthenticatorTypeOOB
			default:
				panic(fmt.Sprintf("interaction: unknown edge: %T", edge))
			}
		},
	)

	return edges
}
