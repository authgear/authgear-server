package interaction

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func (p *Provider) GetInteractionState(i *Interaction) (*State, error) {
	switch intent := i.Intent.(type) {
	case *IntentLogin:
		return p.getStateLogin(i, intent)
	case *IntentSignup:
		return p.getStateSignup(i, intent)
	}
	panic(fmt.Sprintf("interaction: unknown intent type %T", i.Intent))
}

func (p *Provider) getStateLogin(i *Interaction, intent *IntentLogin) (*State, error) {
	secondaryAuthenticators, err := p.listSecondaryAuthenticators(i.UserID)
	if err != nil {
		return nil, err
	}
	primaryAuthenticators := p.getAvailablePrimaryAuthenticators(intent.Identity.Type)
	s := &State{}

	// Primary authentication
	needPrimaryAuthn := false
	if len(primaryAuthenticators) > 0 {
		s.Steps = []StepState{
			{
				Step:                    StepAuthenticatePrimary,
				AvailableAuthenticators: primaryAuthenticators,
			},
		}
		needPrimaryAuthn = true
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
	primaryAuthenticators := p.getAvailablePrimaryAuthenticators(intent.Identity.Type)
	s := &State{}

	// Setup primary authenticator
	needPrimaryAuthn := false
	if len(primaryAuthenticators) > 0 {
		s.Steps = []StepState{
			{
				Step:                    StepAuthenticatePrimary,
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

var identityPrimaryAuthenticators = map[authn.IdentityType]map[AuthenticatorType]bool{
	authn.IdentityTypeLoginID: {
		AuthenticatorTypePassword: true,
		AuthenticatorTypeTOTP:     true,
		AuthenticatorTypeOOBOTP:   true,
	},
}

func (p *Provider) getAvailablePrimaryAuthenticators(typ authn.IdentityType) []AuthenticatorSpec {
	var as []AuthenticatorSpec
	for _, t := range p.Config.PrimaryAuthenticators {
		authenticatorType := AuthenticatorType(t)
		if !identityPrimaryAuthenticators[typ][authenticatorType] {
			continue
		}
		as = append(as, AuthenticatorSpec{Type: authenticatorType, Props: map[string]interface{}{}})
	}
	return as
}

func (p *Provider) getAvailableSecondaryAuthenticators() []AuthenticatorSpec {
	var as []AuthenticatorSpec
	for _, t := range p.Config.SecondaryAuthenticators {
		as = append(as, AuthenticatorSpec{Type: AuthenticatorType(t), Props: map[string]interface{}{}})
	}
	return as
}

func (p *Provider) listSecondaryAuthenticators(userID string) ([]AuthenticatorSpec, error) {
	var as []AuthenticatorSpec
	for _, t := range p.Config.SecondaryAuthenticators {
		ais, err := p.Authenticator.List(userID, AuthenticatorType(t))
		if err != nil {
			return nil, err
		}
		for _, ai := range ais {
			as = append(as, ai.ToSpec())
		}
	}
	return as, nil
}
