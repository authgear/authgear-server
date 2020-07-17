package interaction

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/core/errors"
)

func (p *Provider) GetStepState(i *Interaction) (*StepState, error) {
	var steps []StepState
	var err error
	switch intent := i.Intent.(type) {
	case *IntentOAuth:
		steps, err = p.getStateOAuth(i, intent)
	case *IntentLogin:
		steps, err = p.getStateLogin(i, intent)
	case *IntentSignup:
		steps, err = p.getStateSignup(i, intent)
	case *IntentAddIdentity:
		steps, err = p.getStateAddIdentity(i, intent)
	case *IntentRemoveIdentity:
		steps, err = p.getStateRemoveIdentity(i, intent)
	case *IntentUpdateIdentity:
		steps, err = p.getStateUpdateIdentity(i, intent)
	case *IntentUpdateAuthenticator:
		steps, err = p.getStateUpdateAuthenticator(i, intent)
	default:
		panic(fmt.Sprintf("interaction: unknown intent type %T", i.Intent))
	}
	if err != nil {
		return nil, err
	}
	return &steps[len(steps)-1], nil
}

func (p *Provider) getStateOAuth(i *Interaction, intent *IntentOAuth) (steps []StepState, err error) {
	steps = append(steps, StepState{
		Step:                          StepOAuth,
		Identity:                      intent.Identity,
		OAuthAction:                   intent.Action,
		OAuthNonce:                    intent.Nonce,
		OAuthProviderAuthorizationURL: intent.ProviderAuthorizationURL,
		OAuthUserID:                   intent.UserID,
	})
	return
}

func (p *Provider) getStateLogin(i *Interaction, intent *IntentLogin) (steps []StepState, err error) {
	// Whether or not primary authentication is needed or not
	// solely depends on the identity being used.
	// If for some reason, the user uses login ID Identity but
	// they do not have any authenticator, we must not skip primary authentication.
	needPrimaryAuthn := len(identityPrimaryAuthenticators[intent.Identity.Type]) > 0

	primaryAuthenticators, err := p.listPrimaryAuthenticators(intent.Identity)
	if err != nil {
		return
	}
	secondaryAuthenticators, err := p.listSecondaryAuthenticators(i.UserID)
	if err != nil {
		return
	}

	// Primary authentication
	if needPrimaryAuthn {
		steps = append(steps, StepState{
			Step:                    StepAuthenticatePrimary,
			AvailableAuthenticators: primaryAuthenticators,
		})
	}
	if needPrimaryAuthn && i.PrimaryAuthenticator == nil {
		return
	}

	// Secondary authentication
	needSecondaryAuthn := false
	switch {
	case len(secondaryAuthenticators) > 0 &&
		(p.Config.SecondaryAuthenticationMode == config.SecondaryAuthenticationModeIfExists ||
			p.Config.SecondaryAuthenticationMode == config.SecondaryAuthenticationModeRequired):
		steps = append(steps, StepState{
			Step:                    StepAuthenticateSecondary,
			AvailableAuthenticators: secondaryAuthenticators,
		})
		needSecondaryAuthn = true

	case len(secondaryAuthenticators) == 0 &&
		p.Config.SecondaryAuthenticationMode == config.SecondaryAuthenticationModeRequired:
		steps = append(steps, StepState{
			Step:                    StepSetupSecondaryAuthenticator,
			AvailableAuthenticators: p.getAvailableSecondaryAuthenticators(),
		})
		needSecondaryAuthn = true
	}
	if needSecondaryAuthn && i.SecondaryAuthenticator == nil {
		return
	}

	// Commit
	steps = append(steps, StepState{Step: StepCommit})
	return
}

func (p *Provider) getStateSignup(i *Interaction, intent *IntentSignup) (steps []StepState, err error) {
	primaryAuthenticators := p.getAvailablePrimaryAuthenticators(i.NewIdentities[0])

	// Setup primary authenticator
	needPrimaryAuthn := false
	if len(primaryAuthenticators) > 0 {
		steps = append(steps, StepState{
			Step:                    StepSetupPrimaryAuthenticator,
			AvailableAuthenticators: primaryAuthenticators,
		})
		needPrimaryAuthn = true
	}
	if needPrimaryAuthn && i.PrimaryAuthenticator == nil {
		return
	}

	// Commit
	steps = append(steps, StepState{Step: StepCommit})
	return
}

func (p *Provider) getStateAddIdentity(i *Interaction, intent *IntentAddIdentity) (steps []StepState, err error) {
	if len(i.NewIdentities) != 1 {
		panic("interaction: unexpected number of new identities")
	}

	// check if new authenticator is needed for new identity
	needSetupPrimaryAuthenticators, err := p.getNeedSetupPrimaryAuthenticatorsWithNewIdentity(i.UserID, i.NewIdentities[0])
	if err != nil {
		return
	}
	if len(needSetupPrimaryAuthenticators) > 0 {
		steps = append(steps, StepState{
			Step:                    StepSetupPrimaryAuthenticator,
			AvailableAuthenticators: needSetupPrimaryAuthenticators,
		})
		if i.PrimaryAuthenticator == nil {
			return
		}
	}

	steps = append(steps, StepState{Step: StepCommit})
	return
}

func (p *Provider) getStateRemoveIdentity(i *Interaction, intent *IntentRemoveIdentity) (steps []StepState, err error) {
	if len(i.RemoveIdentities) != 1 {
		panic("interaction: unexpected number of identities to be removed")
	}
	steps = append(steps, StepState{Step: StepCommit})
	return
}

func (p *Provider) getStateUpdateIdentity(i *Interaction, intent *IntentUpdateIdentity) (steps []StepState, err error) {
	if len(i.UpdateIdentities) != 1 {
		panic("interaction: unexpected number of identities to be updated")
	}

	// check if new authenticator is needed for updated identity
	needSetupPrimaryAuthenticators, err := p.getNeedSetupPrimaryAuthenticatorsWithNewIdentity(i.UserID, i.UpdateIdentities[0])
	if err != nil {
		return
	}
	if len(needSetupPrimaryAuthenticators) > 0 {
		steps = append(steps, StepState{
			Step:                    StepSetupPrimaryAuthenticator,
			AvailableAuthenticators: needSetupPrimaryAuthenticators,
		})
		if i.PrimaryAuthenticator == nil {
			return
		}
	}

	steps = append(steps, StepState{Step: StepCommit})
	return
}

func (p *Provider) getStateUpdateAuthenticator(i *Interaction, intent *IntentUpdateAuthenticator) (steps []StepState, err error) {
	// only password authenticator support update
	if intent.Authenticator.Type != authn.AuthenticatorTypePassword {
		panic("interaction: unexpected authenticator type for update " + intent.Authenticator.Type)
	}
	availableAuthenticators := []authenticator.Spec{
		authenticator.Spec{
			Type:  intent.Authenticator.Type,
			Props: map[string]interface{}{},
		},
	}
	steps = append(steps, StepState{
		Step:                    StepSetupPrimaryAuthenticator,
		AvailableAuthenticators: availableAuthenticators,
	})
	if i.PrimaryAuthenticator == nil {
		return
	}

	steps = append(steps, StepState{Step: StepCommit})
	return
}

var identityPrimaryAuthenticators = map[authn.IdentityType]map[authn.AuthenticatorType]bool{
	authn.IdentityTypeLoginID: {
		authn.AuthenticatorTypePassword: true,
		authn.AuthenticatorTypeTOTP:     true,
		authn.AuthenticatorTypeOOB:      true,
	},
}

func (p *Provider) getAvailablePrimaryAuthenticators(ii *identity.Info) []authenticator.Spec {
	var as []authenticator.Spec
	for _, t := range p.Config.PrimaryAuthenticators {
		authenticatorType := t
		if !identityPrimaryAuthenticators[ii.Type][authenticatorType] {
			continue
		}
		spec := p.Identity.RelateIdentityToAuthenticator(ii, &authenticator.Spec{
			Type:  authenticatorType,
			Props: map[string]interface{}{},
		})
		if spec != nil {
			as = append(as, *spec)
		}
	}
	return as
}

func (p *Provider) getAvailableSecondaryAuthenticators() []authenticator.Spec {
	var as []authenticator.Spec
	for _, t := range p.Config.SecondaryAuthenticators {
		as = append(as, authenticator.Spec{Type: authn.AuthenticatorType(t), Props: map[string]interface{}{}})
	}
	return as
}

func (p *Provider) listPrimaryAuthenticators(is identity.Spec) (specs []authenticator.Spec, err error) {
	// Now we use authgear claims to find exactly one identity.
	// In the future we may use OIDC claims to list all identities and
	// resolve which user the actor want to authenticate as.
	userID, ii, err := p.Identity.GetByClaims(is.Type, is.Claims)
	if errors.Is(err, identity.ErrIdentityNotFound) {
		// Now we eagerly check if the identity exists or not.
		err = ErrInvalidCredentials
		return
	} else if err != nil {
		return
	}

	ais, err := p.Authenticator.ListByIdentity(userID, ii)
	if err != nil {
		return
	}

	for _, ai := range ais {
		for _, t := range p.Config.PrimaryAuthenticators {
			if ai.Type == t {
				specs = append(specs, ai.ToSpec())
			}
		}
	}

	return
}

func (p *Provider) listSecondaryAuthenticators(userID string) ([]authenticator.Spec, error) {
	var as []authenticator.Spec
	for _, t := range p.Config.SecondaryAuthenticators {
		ais, err := p.Authenticator.List(userID, authn.AuthenticatorType(t))
		if err != nil {
			return nil, err
		}
		for _, ai := range ais {
			as = append(as, ai.ToSpec())
		}
	}
	return as, nil
}

func (p *Provider) getNeedSetupPrimaryAuthenticatorsWithNewIdentity(userID string, ii *identity.Info) ([]authenticator.Spec, error) {
	availableAuthenticators := p.getAvailablePrimaryAuthenticators(ii)
	identityAuthenticators, err := p.Authenticator.ListByIdentity(userID, ii)
	if err != nil {
		return nil, err
	}

	found := false
	for _, as := range availableAuthenticators {
		for _, ia := range identityAuthenticators {
			if as.Type == ia.Type {
				found = true
			}
		}
	}

	needPrimaryAuthn := len(availableAuthenticators) > 0 && !found
	if needPrimaryAuthn {
		return availableAuthenticators, nil
	}
	return []authenticator.Spec{}, nil
}
