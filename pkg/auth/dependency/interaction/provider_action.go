package interaction

import (
	"errors"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/oob"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity/loginid"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/otp"
)

func (p *Provider) PerformAction(i *Interaction, step Step, action Action) error {
	stepState, err := p.GetStepState(i)
	if err != nil {
		return err
	}

	if stepState.Step != step {
		return ErrInvalidStep
	}

	switch intent := i.Intent.(type) {
	case *IntentLogin:
		return p.performActionLogin(i, intent, stepState, action)
	case *IntentSignup:
		return p.performActionSignup(i, intent, stepState, action)
	case *IntentAddIdentity:
		return p.performActionAddIdentity(i, intent, stepState, action)
	case *IntentUpdateIdentity:
		return p.performActionUpdateIdentity(i, intent, stepState, action)
	case *IntentUpdateAuthenticator:
		return p.performActionUpdateAuthenticator(i, intent, stepState, action)
	}
	panic(fmt.Sprintf("interaction: unknown intent type %T", i.Intent))
}

func (p *Provider) performActionLogin(i *Interaction, intent *IntentLogin, step *StepState, action Action) error {
	switch step.Step {
	case StepAuthenticatePrimary, StepAuthenticateSecondary:
		switch action := action.(type) {
		case *ActionAuthenticate:
			authen, err := p.doAuthenticate(i, step, &i.State, intent.Identity, action.Authenticator, action.Secret)
			if err != nil {
				return err
			}

			ar := authen.ToRef()
			if step.Step == StepAuthenticatePrimary {
				i.PrimaryAuthenticator = &ar
				i.SecondaryAuthenticator = nil
			} else {
				i.SecondaryAuthenticator = &ar
			}
			return nil

		case *ActionTriggerOOBAuthenticator:
			err := p.doTriggerOOB(i, step, action)
			if err != nil {
				return err
			}
			return nil
		default:
			panic(fmt.Sprintf("interaction_login: unhandled authenticate action %T", action))
		}

	case StepSetupSecondaryAuthenticator:
		// TODO(interaction): setup secondary authenticator

	case StepCommit:
		// TODO(interaction): allow setup bearer token

	}
	panic("interaction_login: unhandled step " + step.Step)
}

func (p *Provider) performActionSignup(i *Interaction, intent *IntentSignup, step *StepState, action Action) error {
	switch step.Step {
	case StepSetupPrimaryAuthenticator:
		return p.setupPrimaryAuthenticator(i, step, action)
	}
	panic("interaction_signup: unhandled step " + step.Step)
}

func (p *Provider) performActionAddIdentity(i *Interaction, intent *IntentAddIdentity, step *StepState, action Action) error {
	switch step.Step {
	case StepSetupPrimaryAuthenticator:
		return p.setupPrimaryAuthenticator(i, step, action)
	}
	panic("interaction_add_identity: unhandled step " + step.Step)
}

func (p *Provider) performActionUpdateIdentity(i *Interaction, intent *IntentUpdateIdentity, step *StepState, action Action) error {
	switch step.Step {
	case StepSetupPrimaryAuthenticator:
		// setup primary authenticator for updated identity
		return p.setupPrimaryAuthenticator(i, step, action)
	}
	panic("interaction_add_identity: unhandled step " + step.Step)
}

func (p *Provider) performActionUpdateAuthenticator(i *Interaction, intent *IntentUpdateAuthenticator, step *StepState, action Action) error {
	if step.Step != StepSetupPrimaryAuthenticator {
		panic("interaction_update_authenticator: expected step " + step.Step)
	}

	act, ok := action.(*ActionSetupAuthenticator)
	if !ok {
		panic("interaction_update_authenticator: expected action type")
	}

	ai := i.PendingAuthenticator
	changed, newAuthen, err := p.Authenticator.WithSecret(i.UserID, ai, act.Secret)
	if err != nil {
		return err
	}

	// Add authenticator to UpdateAuthenticators if it is changed only
	if changed {
		i.UpdateAuthenticators = append(i.UpdateAuthenticators, newAuthen)
	}

	ar := newAuthen.ToRef()
	i.PrimaryAuthenticator = &ar
	i.PendingAuthenticator = nil
	return nil

}

func (p *Provider) setupPrimaryAuthenticator(i *Interaction, step *StepState, action Action) error {
	switch action := action.(type) {
	case *ActionSetupAuthenticator:
		authen, err := p.setupAuthenticator(i, step, &i.State, action.Authenticator, action.Secret)
		if err != nil {
			return err
		}

		ar := authen.ToRef()
		i.PrimaryAuthenticator = &ar
		return nil

	case *ActionTriggerOOBAuthenticator:
		err := p.doTriggerOOB(i, step, action)
		if err != nil {
			return err
		}
		return nil
	default:
		panic(fmt.Sprintf("interaction_signup: unhandled authenticate action %T", action))
	}
}

func (p *Provider) doAuthenticate(i *Interaction, step *StepState, astate *map[string]string, is identity.Spec, as authenticator.Spec, secret string) (*authenticator.Info, error) {
	userID, iden, err := p.Identity.GetByClaims(is.Type, is.Claims)
	if errors.Is(err, identity.ErrIdentityNotFound) {
		return nil, ErrInvalidCredentials
	} else if err != nil {
		return nil, err
	}

	authen, err := p.Authenticator.Authenticate(userID, as, astate, secret)
	if err != nil {
		return nil, err
	}

	i.UserID = userID
	ir := iden.ToRef()
	i.Identity = &ir
	i.State = map[string]string{}
	return authen, nil
}

func (p *Provider) setupAuthenticator(i *Interaction, step *StepState, astate *map[string]string, as authenticator.Spec, secret string) (*authenticator.Info, error) {
	ok := false
	for _, aa := range step.AvailableAuthenticators {
		if aa.Type == as.Type {
			ok = true
			break
		}
	}
	if !ok {
		// Authenticator is not available for current step, reject it
		return nil, ErrInvalidAction
	}

	switch as.Type {
	case authn.AuthenticatorTypePassword:
		// Nothing special needs to be done
		break
	case authn.AuthenticatorTypeOOB:
		// Ignoring the first return value because it is always nil.
		_, err := p.Authenticator.Authenticate(i.UserID, as, astate, secret)
		if err != nil {
			return nil, err
		}
	default:
		panic("interaction_signup: setup up unexpected authenticator type: " + as.Type)
	}

	ais, err := p.Authenticator.New(i.UserID, as, secret)
	if err != nil {
		return nil, err
	}
	i.NewAuthenticators = append(i.NewAuthenticators, ais...)
	i.State = map[string]string{}
	return ais[0], nil
}

func (p *Provider) doTriggerOOB(i *Interaction, step *StepState, action *ActionTriggerOOBAuthenticator) (err error) {
	spec := action.Authenticator

	if spec.Type != authn.AuthenticatorTypeOOB {
		panic("interaction: unexpected ActionTriggerOOBAuthenticator.Authenticator.Type: " + spec.Type)
	}

	now := p.Clock.NowUTC()
	nowBytes, err := now.MarshalText()
	if err != nil {
		return
	}
	nowStr := string(nowBytes)

	if i.State == nil {
		i.State = map[string]string{}
	}

	// Rotate the code according to oob.OOBCodeValidDuration
	code := i.State[authenticator.AuthenticatorStateOOBOTPCode]
	generateTimeStr := i.State[authenticator.AuthenticatorStateOOBOTPGenerateTime]
	channel := spec.Props[authenticator.AuthenticatorPropOOBOTPChannelType].(string)
	if generateTimeStr == "" {
		code = p.OOB.GenerateCode(authn.AuthenticatorOOBChannel(channel))
		generateTimeStr = nowStr
	} else {
		var tt time.Time
		err = tt.UnmarshalText([]byte(generateTimeStr))
		if err != nil {
			return
		}

		// Expire
		if tt.Add(oob.OOBOTPValidDuration).Before(now) {
			code = p.OOB.GenerateCode(authn.AuthenticatorOOBChannel(channel))
			generateTimeStr = nowStr
		}
	}

	// Respect cooldown
	triggerTimeStr := i.State[authenticator.AuthenticatorStateOOBOTPTriggerTime]
	if triggerTimeStr != "" {
		var tt time.Time
		err = tt.UnmarshalText([]byte(triggerTimeStr))
		if err != nil {
			return
		}

		if tt.Add(oob.OOBOTPSendCooldownSeconds * time.Second).After(now) {
			err = ErrOOBOTPCooldown
			return
		}
	}

	err = p.sendOOBCode(i, step, spec, code)
	if err != nil {
		return
	}

	// Perform mutation on interaction at the end.

	// This function can be called by login or signup.
	// In case of signup, the spec does not have an ID yet.
	if id, ok := spec.Props[authenticator.AuthenticatorPropOOBOTPID].(string); ok {
		i.State[authenticator.AuthenticatorStateOOBOTPID] = id
	}
	i.State[authenticator.AuthenticatorStateOOBOTPCode] = code
	i.State[authenticator.AuthenticatorStateOOBOTPGenerateTime] = generateTimeStr
	i.State[authenticator.AuthenticatorStateOOBOTPTriggerTime] = nowStr
	i.State[authenticator.AuthenticatorStateOOBOTPChannelType] = channel

	return
}

func (p *Provider) sendOOBCode(i *Interaction, step *StepState, as authenticator.Spec, code string) error {
	channel, ok := as.Props[authenticator.AuthenticatorPropOOBOTPChannelType].(string)
	if !ok {
		panic("interaction: cannot extract authenticator channel")
	}

	var operation otp.OOBOperationType
	var loginID *loginid.LoginID
	switch step.Step {
	case StepSetupPrimaryAuthenticator, StepAuthenticatePrimary:
		operation = otp.OOBOperationTypePrimaryAuth
		// Primary OOB authenticators is bound to login ID identities:
		// Extract login ID from the bound identity.

		identityID, ok := as.Props[authenticator.AuthenticatorPropOOBOTPIdentityID].(*string)
		if !ok || identityID == nil {
			// No bound identity found: break and use placeholder login ID
			break
		}
		var boundIdentity *identity.Info
		for _, iden := range i.NewIdentities {
			if iden.ID == *identityID {
				boundIdentity = iden
				break
			}
		}
		if boundIdentity == nil {
			for _, iden := range i.UpdateIdentities {
				if iden.ID == *identityID {
					boundIdentity = iden
					break
				}
			}
		}

		if boundIdentity == nil {
			var err error
			boundIdentity, err = p.Identity.Get(i.UserID, authn.IdentityTypeLoginID, *identityID)
			if errors.Is(err, identity.ErrIdentityNotFound) {
				// No bound identity found: break and use placeholder login ID
				break
			} else if err != nil {
				return err
			}
		}

		loginID = &loginid.LoginID{
			Key:   boundIdentity.Claims[identity.IdentityClaimLoginIDKey].(string),
			Value: boundIdentity.Claims[identity.IdentityClaimLoginIDValue].(string),
		}

	case StepSetupSecondaryAuthenticator, StepAuthenticateSecondary:
		operation = otp.OOBOperationTypeSecondaryAuth
		// Secondary OOB authenticators is not bound to login ID identities.
		loginID = nil

	default:
		panic("interaction: attempted to trigger OOB in unexpected step: " + step.Step)
	}

	var origin otp.MessageOrigin
	switch i.Intent.Type() {
	case IntentTypeLogin:
		origin = otp.MessageOriginLogin
	case IntentTypeSignup:
		origin = otp.MessageOriginSignup
	default:
		origin = otp.MessageOriginSettings
	}

	if loginID == nil {
		// Use a placeholder login ID if no bound login ID identity
		loginID = &loginid.LoginID{}
		switch channel {
		case string(authn.AuthenticatorOOBChannelSMS):
			loginID.Value = as.Props[authenticator.AuthenticatorPropOOBOTPPhone].(string)
		case string(authn.AuthenticatorOOBChannelEmail):
			loginID.Value = as.Props[authenticator.AuthenticatorPropOOBOTPEmail].(string)
		}
	}

	return p.OOB.SendCode(
		authn.AuthenticatorOOBChannel(channel),
		loginID,
		code,
		origin,
		operation,
	)
}
