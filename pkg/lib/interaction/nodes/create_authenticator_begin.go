package nodes

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func init() {
	interaction.RegisterNode(&NodeCreateAuthenticatorBegin{})
}

type EdgeCreateAuthenticatorBegin struct {
	Stage             authn.AuthenticationStage
	AuthenticatorType *model.AuthenticatorType
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
		NewAuthenticatorID: uuid.New(),
		Stage:              e.Stage,
		AuthenticatorType:  e.AuthenticatorType,
		SkipMFASetup:       skipMFASetup,
		RequestedByUser:    requestedByUser,
	}, nil
}

type NodeCreateAuthenticatorBegin struct {
	NewAuthenticatorID string                    `json:"new_authenticator_id"`
	Stage              authn.AuthenticationStage `json:"stage"`
	AuthenticatorType  *model.AuthenticatorType  `json:"authenticator_type"`
	SkipMFASetup       bool                      `json:"skip_mfa_setup"`
	RequestedByUser    bool                      `json:"requested_by_user"`

	Identity             *identity.Info               `json:"-"`
	PrimaryAuthenticator *authenticator.Info          `json:"-"`
	AuthenticationConfig *config.AuthenticationConfig `json:"-"`
	AuthenticatorConfig  *config.AuthenticatorConfig  `json:"-"`
	Authenticators       []*authenticator.Info        `json:"-"`
}

func (n *NodeCreateAuthenticatorBegin) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	ais, err := ctx.Authenticators.List(graph.MustGetUserID())
	if err != nil {
		return err
	}

	if iden, ok := graph.GetUserLastIdentity(); ok {
		n.Identity = iden
	}
	if prim, ok := graph.GetUserAuthenticator(authn.AuthenticationStagePrimary); ok {
		n.PrimaryAuthenticator = prim
	}
	n.AuthenticationConfig = ctx.Config.Authentication
	n.AuthenticatorConfig = ctx.Config.Authenticator
	n.Authenticators = ais
	return nil
}

func (n *NodeCreateAuthenticatorBegin) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeCreateAuthenticatorBegin) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return n.deriveEdges()
}

// IsOOBAuthenticatorTypeAllowed implements SetupOOBOTPNode.
func (n *NodeCreateAuthenticatorBegin) IsOOBAuthenticatorTypeAllowed(oobAuthenticatorType model.AuthenticatorType) (bool, error) {
	edges, err := n.deriveEdges()
	if err != nil {
		return false, err
	}

	for _, edge := range edges {
		switch edge := edge.(type) {
		case *EdgeCreateAuthenticatorOOBSetup:
			if edge.OOBAuthenticatorType == oobAuthenticatorType {
				return true, nil
			}
		}
	}

	return false, nil
}

// GetCreateAuthenticatorEdges implements CreateAuthenticatorBeginNode.
func (n *NodeCreateAuthenticatorBegin) GetCreateAuthenticatorEdges() ([]interaction.Edge, error) {
	return n.deriveEdges()
}

func (n *NodeCreateAuthenticatorBegin) GetCreateAuthenticatorStage() authn.AuthenticationStage {
	return n.Stage
}

func (n *NodeCreateAuthenticatorBegin) deriveEdges() ([]interaction.Edge, error) {
	var edges []interaction.Edge
	var err error

	switch n.Stage {
	case authn.AuthenticationStagePrimary:
		if n.AuthenticatorType == nil {
			edges, err = n.derivePrimary()
			if err != nil {
				return nil, err
			}
		} else {
			edges, err = n.derivePrimaryWithType(*n.AuthenticatorType)
			if err != nil {
				return nil, err
			}
		}
	case authn.AuthenticationStageSecondary:
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

func (n *NodeCreateAuthenticatorBegin) derivePrimaryWithType(typ model.AuthenticatorType) ([]interaction.Edge, error) {
	// derivePrimary() assumes the presence of n.Identity
	// n.Identity is absent when we are adding passkey in settings.
	types := *n.AuthenticationConfig.PrimaryAuthenticators

	var edges []interaction.Edge
	for _, t := range types {
		if t == typ {
			if t == model.AuthenticatorTypePasskey {
				edges = append(edges, &EdgeCreateAuthenticatorPasskey{
					NewAuthenticatorID: n.NewAuthenticatorID,
					Stage:              n.Stage,
					IsDefault:          false,
				})
			}
		}
	}

	interaction.SortAuthenticators(
		types,
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

	return edges, nil
}

// nolint:gocognit
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
	types := *n.AuthenticationConfig.PrimaryAuthenticators
	if len(types) == 0 {
		return nil, fmt.Errorf("identity requires primary authenticator but none is enabled")
	}

	// 3. Find out whether the identity has the preferred primary authenticator.
	// If it does not, creation is needed.
	ais := authenticator.ApplyFilters(
		n.Authenticators,
		authenticator.KeepType(types...),
		authenticator.KeepPrimaryAuthenticatorOfIdentity(n.Identity),
	)
	if len(ais) != 0 {
		return nil, nil
	}

	// Primary authenticator is default if it is the first primary authenticator of the user.
	isDefault := len(authenticator.ApplyFilters(n.Authenticators, authenticator.KeepKind(authenticator.KindPrimary))) == 0

	var edges []interaction.Edge
	for _, t := range types {
		switch t {
		case model.AuthenticatorTypePassword:
			edges = append(edges, &EdgeCreateAuthenticatorPassword{
				NewAuthenticatorID: n.NewAuthenticatorID,
				Stage:              n.Stage,
				IsDefault:          isDefault,
			})

		case model.AuthenticatorTypePasskey:
			edges = append(edges, &EdgeCreateAuthenticatorPasskey{
				NewAuthenticatorID: n.NewAuthenticatorID,
				Stage:              n.Stage,
				IsDefault:          isDefault,
			})

		case model.AuthenticatorTypeOOBSMS:
			// check if identity login id type match oob type
			if n.Identity.LoginID != nil {
				if n.Identity.LoginID.LoginIDType == model.LoginIDKeyTypePhone {
					if n.AuthenticatorConfig.OOB.SMS.PhoneOTPMode.IsWhatsappEnabled() {
						edges = append(edges, &EdgeCreateAuthenticatorWhatsappOTPSetup{
							NewAuthenticatorID: n.NewAuthenticatorID,
							Stage:              n.Stage,
							IsDefault:          isDefault,
						})
					}

					if n.AuthenticatorConfig.OOB.SMS.PhoneOTPMode.IsSMSEnabled() {
						edges = append(edges, &EdgeCreateAuthenticatorOOBSetup{
							NewAuthenticatorID:   n.NewAuthenticatorID,
							Stage:                n.Stage,
							IsDefault:            isDefault,
							OOBAuthenticatorType: model.AuthenticatorTypeOOBSMS,
						})
					}
				}
			}

		case model.AuthenticatorTypeOOBEmail:
			// check if identity login id type match oob type
			if n.Identity.LoginID != nil {
				if n.Identity.LoginID.LoginIDType == model.LoginIDKeyTypeEmail {
					if n.AuthenticatorConfig.OOB.Email.EmailOTPMode.IsLoginLinkEnabled() {
						edges = append(edges, &EdgeCreateAuthenticatorLoginLinkOTPSetup{
							NewAuthenticatorID: n.NewAuthenticatorID,
							Stage:              n.Stage,
							IsDefault:          isDefault,
						})
					}

					if n.AuthenticatorConfig.OOB.Email.EmailOTPMode.IsCodeEnabled() {
						edges = append(edges, &EdgeCreateAuthenticatorOOBSetup{
							NewAuthenticatorID:   n.NewAuthenticatorID,
							Stage:                n.Stage,
							IsDefault:            isDefault,
							OOBAuthenticatorType: model.AuthenticatorTypeOOBEmail,
						})
					}
				}
			}

		default:
			panic(fmt.Sprintf("interaction: unknown authenticator type: %s", t))
		}
	}

	if len(edges) == 0 {
		// A new authenticator is required, but no authenticator can be created:
		// Configuration is invalid.
		return nil, fmt.Errorf("no primary authenticator can be created for identity")
	}

	interaction.SortAuthenticators(
		types,
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

	return edges, nil
}

// nolint: gocognit
func (n *NodeCreateAuthenticatorBegin) deriveSecondary() (edges []interaction.Edge) {
	// Determine whether we need to create secondary authenticator.

	ais := authenticator.ApplyFilters(
		n.Authenticators,
		authenticator.KeepKind(authenticator.KindSecondary),
	)

	// Skip setup if explicitly requested
	if n.SkipMFASetup {
		return nil
	}

	// Skip setup if MFA is disabled
	mode := n.AuthenticationConfig.SecondaryAuthenticationMode
	if mode.IsDisabled() {
		return nil
	}

	if !n.RequestedByUser {
		// Skip setup if the primary authenticator being used cannot use MFA.
		if n.PrimaryAuthenticator == nil || !n.PrimaryAuthenticator.CanHaveMFA() {
			return nil
		}

		// Check secondary authentication mode.
		switch mode {
		case config.SecondaryAuthenticationModeIfExists:
			// secondary authentication is optional.
			return nil
		case config.SecondaryAuthenticationModeRequired:
			// secondary authentication is required.
			// Skip setup if the user has at least one secondary authenticator.
			if len(ais) > 0 {
				return nil
			}
		}
	}

	// The created authenticator is default if no other default authenticator
	// exists
	isDefault := len(authenticator.ApplyFilters(ais, authenticator.KeepDefault)) == 0

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
		case model.AuthenticatorTypePassword:
			passwordCount++
		case model.AuthenticatorTypeTOTP:
			totpCount++
		case model.AuthenticatorTypeOOBEmail:
			oobEmailCount++
		case model.AuthenticatorTypeOOBSMS:
			oobSMSCount++
		default:
			panic("interaction: unknown authenticator type: " + a.Type)
		}
	}

	// Condition A.
	for _, typ := range *n.AuthenticationConfig.SecondaryAuthenticators {
		switch typ {
		case model.AuthenticatorTypePassword:
			// Condition B.
			edges = append(edges, &EdgeCreateAuthenticatorPassword{
				NewAuthenticatorID: n.NewAuthenticatorID,
				Stage:              n.Stage,
				IsDefault:          isDefault,
			})
		case model.AuthenticatorTypeTOTP:
			// Condition B and C.
			if totpCount < *n.AuthenticatorConfig.TOTP.Maximum {
				edges = append(edges, &EdgeCreateAuthenticatorTOTPSetup{
					NewAuthenticatorID: n.NewAuthenticatorID,
					Stage:              n.Stage,
					IsDefault:          isDefault,
				})
			}
		case model.AuthenticatorTypeOOBEmail:
			// Condition B and C.
			if oobEmailCount < *n.AuthenticatorConfig.OOB.Email.Maximum {
				if n.AuthenticatorConfig.OOB.Email.EmailOTPMode.IsCodeEnabled() {
					edges = append(edges, &EdgeCreateAuthenticatorOOBSetup{
						NewAuthenticatorID:   n.NewAuthenticatorID,
						Stage:                n.Stage,
						IsDefault:            isDefault,
						OOBAuthenticatorType: model.AuthenticatorTypeOOBEmail,
					})
				}

				if n.AuthenticatorConfig.OOB.Email.EmailOTPMode.IsLoginLinkEnabled() {
					edges = append(edges, &EdgeCreateAuthenticatorLoginLinkOTPSetup{
						NewAuthenticatorID: n.NewAuthenticatorID,
						Stage:              n.Stage,
						IsDefault:          isDefault,
					})
				}
			}
		case model.AuthenticatorTypeOOBSMS:
			// Condition B and C.
			if oobSMSCount < *n.AuthenticatorConfig.OOB.SMS.Maximum {
				if n.AuthenticatorConfig.OOB.SMS.PhoneOTPMode.IsWhatsappEnabled() {
					edges = append(edges, &EdgeCreateAuthenticatorWhatsappOTPSetup{
						NewAuthenticatorID: n.NewAuthenticatorID,
						Stage:              n.Stage,
						IsDefault:          isDefault,
					})
				}

				if n.AuthenticatorConfig.OOB.SMS.PhoneOTPMode.IsSMSEnabled() {
					edges = append(edges, &EdgeCreateAuthenticatorOOBSetup{
						NewAuthenticatorID:   n.NewAuthenticatorID,
						Stage:                n.Stage,
						IsDefault:            isDefault,
						OOBAuthenticatorType: model.AuthenticatorTypeOOBSMS,
					})
				}
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
		*n.AuthenticationConfig.SecondaryAuthenticators,
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
