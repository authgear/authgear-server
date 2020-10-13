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
	interaction.RegisterNode(&NodeCreateAuthenticatorBegin{})
}

type EdgeCreateAuthenticatorBegin struct {
	Stage             interaction.AuthenticationStage
	AuthenticatorType *authn.AuthenticatorType
}

func (e *EdgeCreateAuthenticatorBegin) Instantiate(ctx *interaction.Context, graph *interaction.Graph, input interface{}) (interaction.Node, error) {
	skipMFASetup := false
	var skipMFASetupInput interface{ SkipMFASetup() bool }
	if interaction.Input(input, &skipMFASetupInput) {
		skipMFASetup = skipMFASetupInput.SkipMFASetup()
	}

	requestedByUser := false
	var requestedByUserInput interface{ RequestedByUser() bool }
	if interaction.Input(input, &requestedByUserInput) {
		requestedByUser = requestedByUserInput.RequestedByUser()
	}

	return &NodeCreateAuthenticatorBegin{
		Stage:             e.Stage,
		AuthenticatorType: e.AuthenticatorType,
		SkipMFASetup:      skipMFASetup,
		RequestedByUser:   requestedByUser,
	}, nil
}

type NodeCreateAuthenticatorBegin struct {
	Stage             interaction.AuthenticationStage `json:"stage"`
	AuthenticatorType *authn.AuthenticatorType        `json:"authenticator_type"`
	SkipMFASetup      bool                            `json:"skip_mfa_setup"`
	RequestedByUser   bool                            `json:"requested_by_user"`

	Identity             *identity.Info               `json:"-"`
	AuthenticationConfig *config.AuthenticationConfig `json:"-"`
	AuthenticatorConfig  *config.AuthenticatorConfig  `json:"-"`
	Authenticators       []*authenticator.Info        `json:"-"`
}

func (n *NodeCreateAuthenticatorBegin) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	ais, err := ctx.Authenticators.List(graph.MustGetUserID())
	if err != nil {
		return err
	}

	if n.Stage == interaction.AuthenticationStagePrimary {
		n.Identity = graph.MustGetUserLastIdentity()
	}
	n.AuthenticationConfig = ctx.Config.Authentication
	n.AuthenticatorConfig = ctx.Config.Authenticator
	n.Authenticators = ais
	return nil
}

func (n *NodeCreateAuthenticatorBegin) Apply(perform func(eff interaction.Effect) error, graph *interaction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorBegin) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
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
func (n *NodeCreateAuthenticatorBegin) GetCreateAuthenticatorEdges() ([]interaction.Edge, error) {
	return n.deriveEdges()
}

func (n *NodeCreateAuthenticatorBegin) deriveEdges() ([]interaction.Edge, error) {
	var edges []interaction.Edge
	var err error

	switch n.Stage {
	case interaction.AuthenticationStagePrimary:
		edges, err = n.derivePrimary()
		if err != nil {
			return nil, err
		}
	case interaction.AuthenticationStageSecondary:
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

func (n *NodeCreateAuthenticatorBegin) derivePrimary() ([]interaction.Edge, error) {
	// Determine whether we need to create primary authenticator.

	// 1. Check whether the identity actually requires primary authenticator.
	// If it does not, then no primary authenticator is needed.
	identityRequiresPrimaryAuthentication := len(n.Identity.PrimaryAuthenticatorTypes()) > 0
	if !identityRequiresPrimaryAuthentication {
		return nil, nil
	}

	// 2. Check what primary authenticator the developer prefers.
	// Here we check if the configuration is non-sense.
	types := n.AuthenticationConfig.PrimaryAuthenticators
	if len(types) == 0 {
		return nil, interaction.InvalidConfiguration.New("identity requires primary authenticator but none is enabled")
	}

	// 3. Find out whether the identity has the preferred primary authenticator.
	// If it does not, creation is needed.
	ais := filterAuthenticators(
		n.Authenticators,
		authenticator.KeepType(types...),
		authenticator.KeepPrimaryAuthenticatorOfIdentity(n.Identity),
	)
	if len(ais) != 0 {
		return nil, nil
	}

	// Primary authenticator is default if it is the first primary authenticator of the user.
	isDefault := len(filterAuthenticators(n.Authenticators, authenticator.KeepKind(authenticator.KindPrimary))) == 0

	var edges []interaction.Edge
	for _, t := range types {
		switch t {
		case authn.AuthenticatorTypePassword:
			edges = append(edges, &EdgeCreateAuthenticatorPassword{
				Stage:     n.Stage,
				IsDefault: isDefault,
			})

		case authn.AuthenticatorTypeTOTP:
			edges = append(edges, &EdgeCreateAuthenticatorTOTPSetup{
				Stage:     n.Stage,
				IsDefault: isDefault,
			})

		case authn.AuthenticatorTypeOOB:
			loginIDType := n.Identity.Claims[identity.IdentityClaimLoginIDType].(string)
			loginID := n.Identity.Claims[identity.IdentityClaimLoginIDValue].(string)

			var channel authn.AuthenticatorOOBChannel
			var target string
			switch loginIDType {
			case string(config.LoginIDKeyTypeEmail):
				channel = authn.AuthenticatorOOBChannelEmail
				target = loginID
			case string(config.LoginIDKeyTypePhone):
				channel = authn.AuthenticatorOOBChannelSMS
				target = loginID
			}

			if target != "" {
				edges = append(edges, &EdgeCreateAuthenticatorOOBSetup{
					Stage:     n.Stage,
					IsDefault: isDefault,
					Channel:   channel,
					Target:    target,
				})
			}
		default:
			panic(fmt.Sprintf("interaction: unknown authenticator type: %s", t))
		}
	}

	if len(edges) == 0 {
		// A new authenticator is required, but no authenticator can be created:
		// Configuration is invalid.
		return nil, interaction.InvalidConfiguration.New("no primary authenticator can be created for identity")
	}

	// TODO(interaction): support switching of primary authenticator type to create
	// Return first edge for now.
	return edges[:1], nil
}

func (n *NodeCreateAuthenticatorBegin) deriveSecondary() (edges []interaction.Edge) {
	// Determine whether we need to create secondary authenticator.

	// 1. Skip setup if explicitly requested
	if n.SkipMFASetup {
		return nil
	}

	ais := filterAuthenticators(
		n.Authenticators,
		authenticator.KeepKind(authenticator.KindSecondary),
	)

	// 2. Check secondary authentication mode.
	mode := n.AuthenticationConfig.SecondaryAuthenticationMode
	switch mode {
	case config.SecondaryAuthenticationModeIfRequested:
		// Create only if requested by user
		if !n.RequestedByUser {
			return nil
		}

	case config.SecondaryAuthenticationModeIfExists:
		// Same as IfRequested:
		// Create only if requested by user
		if !n.RequestedByUser {
			return nil
		}

	case config.SecondaryAuthenticationModeRequired:
		// Require at least one secondary authenticator:
		// Skip creation if any secondary authenticator exists and
		// not explicitly requested by user
		if len(ais) > 0 && !n.RequestedByUser {
			return nil
		}
	}

	// The created authenticator is default if no other default authenticator
	// exists
	isDefault := len(filterAuthenticators(ais, authenticator.KeepDefault)) == 0

	// 3. Determine what secondary authenticator we allow the user to create.
	// We have the following conditions to hold:
	//   A. The secondary authenticator is allowed in the configuration.
	//   B. The user does not have that type of secondary authenticator yet. (This is always true since we have 2)
	//   C. The number of the secondary authenticator the user is less than maximum.
	//   D. The secondary authenticator is required by the caller.

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
			channel := a.Claims[authenticator.AuthenticatorClaimOOBOTPChannelType].(string)
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
			// Condition B.
			edges = append(edges, &EdgeCreateAuthenticatorPassword{
				Stage:     n.Stage,
				IsDefault: isDefault,
			})
		case authn.AuthenticatorTypeTOTP:
			// Condition B and C.
			if totpCount < *n.AuthenticatorConfig.TOTP.Maximum {
				edges = append(edges, &EdgeCreateAuthenticatorTOTPSetup{
					Stage:     n.Stage,
					IsDefault: isDefault,
				})
			}
		case authn.AuthenticatorTypeOOB:
			var allowedChannels []authn.AuthenticatorOOBChannel
			// Condition B and C.
			if oobSMSCount < *n.AuthenticatorConfig.OOB.SMS.Maximum {
				allowedChannels = append(allowedChannels, authn.AuthenticatorOOBChannelSMS)
			}
			// Condition B and C.
			if oobEmailCount < *n.AuthenticatorConfig.OOB.Email.Maximum {
				allowedChannels = append(allowedChannels, authn.AuthenticatorOOBChannelEmail)
			}
			if len(allowedChannels) > 0 {
				edges = append(edges, &EdgeCreateAuthenticatorOOBSetup{
					Stage:           n.Stage,
					IsDefault:       isDefault,
					AllowedChannels: allowedChannels,
				})
			}
		default:
			panic("interaction: unknown authenticator type: " + typ)
		}
	}

	// Condition D.
	if n.AuthenticatorType != nil {
		t := *n.AuthenticatorType
		n := 0
		for _, e := range edges {
			edge := e.(interaction.SortableAuthenticator)
			if edge.AuthenticatorType() == t {
				edges[n] = e
				n++
			}
		}
		edges = edges[:n]
	}

	interaction.SortAuthenticators(
		n.AuthenticationConfig.SecondaryAuthenticators,
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

	return edges
}
