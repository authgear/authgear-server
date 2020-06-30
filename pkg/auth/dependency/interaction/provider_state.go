package interaction

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/core/errors"
)

func (p *Provider) GetInteractionState(i *Interaction) (*State, error) {
	switch intent := i.Intent.(type) {
	case *IntentLogin:
		return p.getStateLogin(i, intent)
	case *IntentSignup:
		return p.getStateSignup(i, intent)
	case *IntentAddIdentity:
		return p.getStateAddIdentity(i, intent)
	case *IntentRemoveIdentity:
		return p.getStateRemoveIdentity(i, intent)
	case *IntentUpdateIdentity:
		return p.getStateUpdateIdentity(i, intent)
	case *IntentUpdateAuthenticator:
		return p.getStateUpdateAuthenticator(i, intent)
	}
	panic(fmt.Sprintf("interaction: unknown intent type %T", i.Intent))
}

func (p *Provider) getStateLogin(i *Interaction, intent *IntentLogin) (*State, error) {
	// Whether or not primary authentication is needed or not
	// solely depends on the identity being used.
	// If for some reason, the user uses login ID Identity but
	// they do not have any authenticator, we must not skip primary authentication.
	needPrimaryAuthn := len(identityPrimaryAuthenticators[intent.Identity.Type]) > 0

	primaryAuthenticators, err := p.listPrimaryAuthenticators(intent.Identity)
	if err != nil {
		return nil, err
	}
	secondaryAuthenticators, err := p.listSecondaryAuthenticators(i.UserID)
	if err != nil {
		return nil, err
	}
	s := &State{}

	// Primary authentication
	if needPrimaryAuthn {
		s.Steps = []StepState{
			{
				Step:                    StepAuthenticatePrimary,
				AvailableAuthenticators: primaryAuthenticators,
			},
		}
	}
	if needPrimaryAuthn && i.PrimaryAuthenticator == nil {
		return s, nil
	}

	// Secondary authentication
	needSecondaryAuthn := false
	switch {
	case len(secondaryAuthenticators) > 0 &&
		(p.Config.SecondaryAuthenticationMode == config.SecondaryAuthenticationModeIfExists ||
			p.Config.SecondaryAuthenticationMode == config.SecondaryAuthenticationModeRequired):
		s.Steps = append(s.Steps, StepState{
			Step:                    StepAuthenticateSecondary,
			AvailableAuthenticators: secondaryAuthenticators,
		})
		needSecondaryAuthn = true

	case len(secondaryAuthenticators) == 0 &&
		p.Config.SecondaryAuthenticationMode == config.SecondaryAuthenticationModeRequired:
		s.Steps = append(s.Steps, StepState{
			Step:                    StepSetupSecondaryAuthenticator,
			AvailableAuthenticators: p.getAvailableSecondaryAuthenticators(),
		})
		needSecondaryAuthn = true
	}
	if needSecondaryAuthn && i.SecondaryAuthenticator == nil {
		return s, nil
	}

	// Commit
	s.Steps = append(s.Steps, StepState{Step: StepCommit})
	return s, nil
}

func (p *Provider) getStateSignup(i *Interaction, intent *IntentSignup) (*State, error) {
	primaryAuthenticators := p.getAvailablePrimaryAuthenticators(intent.Identity)
	s := &State{}

	// Setup primary authenticator
	needPrimaryAuthn := false
	if len(primaryAuthenticators) > 0 {
		s.Steps = []StepState{
			{
				Step:                    StepSetupPrimaryAuthenticator,
				AvailableAuthenticators: primaryAuthenticators,
			},
		}
		needPrimaryAuthn = true
	}
	if needPrimaryAuthn && i.PrimaryAuthenticator == nil {
		return s, nil
	}

	// Commit
	s.Steps = append(s.Steps, StepState{Step: StepCommit})
	return s, nil
}

func (p *Provider) getStateAddIdentity(i *Interaction, intent *IntentAddIdentity) (*State, error) {
	s := &State{}
	if len(i.NewIdentities) != 1 {
		panic("interaction: unexpected number of new identities")
	}

	// check if new authenticator is needed for new identity
	needSetupPrimaryAuthenticators, err := p.getNeedSetupPrimaryAuthenticatorsWithNewIdentity(i.UserID, intent.Identity, i.NewIdentities[0])
	if err != nil {
		return nil, err
	}
	if len(needSetupPrimaryAuthenticators) > 0 {
		s.Steps = []StepState{
			{
				Step:                    StepSetupPrimaryAuthenticator,
				AvailableAuthenticators: needSetupPrimaryAuthenticators,
			},
		}
		if i.PrimaryAuthenticator == nil {
			return s, nil
		}
	}

	s.Steps = append(s.Steps, StepState{Step: StepCommit})
	return s, nil
}

func (p *Provider) getStateRemoveIdentity(i *Interaction, intent *IntentRemoveIdentity) (*State, error) {
	s := &State{}
	if len(i.RemoveIdentities) != 1 {
		panic("interaction: unexpected number of identities to be removed")
	}
	s.Steps = append(s.Steps, StepState{Step: StepCommit})
	return s, nil
}

func (p *Provider) getStateUpdateIdentity(i *Interaction, intent *IntentUpdateIdentity) (*State, error) {
	s := &State{}
	if len(i.UpdateIdentities) != 1 {
		panic("interaction: unexpected number of identities to be updated")
	}

	// check if new authenticator is needed for updated identity
	needSetupPrimaryAuthenticators, err := p.getNeedSetupPrimaryAuthenticatorsWithNewIdentity(i.UserID, intent.NewIdentity, i.UpdateIdentities[0])
	if err != nil {
		return nil, err
	}
	if len(needSetupPrimaryAuthenticators) > 0 {
		s.Steps = []StepState{
			{
				Step:                    StepSetupPrimaryAuthenticator,
				AvailableAuthenticators: needSetupPrimaryAuthenticators,
			},
		}
		if i.PrimaryAuthenticator == nil {
			return s, nil
		}
	}

	s.Steps = append(s.Steps, StepState{Step: StepCommit})
	return s, nil
}

func (p *Provider) getStateUpdateAuthenticator(i *Interaction, intent *IntentUpdateAuthenticator) (*State, error) {
	s := &State{}
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
	s.Steps = []StepState{
		{
			Step:                    StepSetupPrimaryAuthenticator,
			AvailableAuthenticators: availableAuthenticators,
		},
	}
	if i.PrimaryAuthenticator == nil {
		return s, nil
	}
	s.Steps = append(s.Steps, StepState{Step: StepCommit})
	return s, nil
}

var identityPrimaryAuthenticators = map[authn.IdentityType]map[authn.AuthenticatorType]bool{
	authn.IdentityTypeLoginID: {
		authn.AuthenticatorTypePassword: true,
		authn.AuthenticatorTypeTOTP:     true,
		authn.AuthenticatorTypeOOB:      true,
	},
}

func (p *Provider) getAvailablePrimaryAuthenticators(is identity.Spec) []authenticator.Spec {
	var as []authenticator.Spec
	for _, t := range p.Config.PrimaryAuthenticators {
		authenticatorType := authn.AuthenticatorType(t)
		if !identityPrimaryAuthenticators[is.Type][authenticatorType] {
			continue
		}
		spec := p.Identity.RelateIdentityToAuthenticator(is, &authenticator.Spec{
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
	// Now we use skygear claims to find exactly one identity.
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

func (p *Provider) getNeedSetupPrimaryAuthenticatorsWithNewIdentity(userID string, is identity.Spec, ii *identity.Info) ([]authenticator.Spec, error) {
	availableAuthenticators := p.getAvailablePrimaryAuthenticators(is)
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
