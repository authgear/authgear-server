package interaction

import (
	"errors"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

func (p *Provider) NewInteractionLoginAs(
	intent *IntentLogin,
	userID string,
	identityRef *IdentityRef,
	primaryAuthenticatorRef *AuthenticatorRef,
	clientID string,
) (*Interaction, error) {
	identity, err := p.Identity.Get(
		userID,
		identityRef.Type,
		identityRef.ID)
	if err != nil {
		return nil, err
	}
	i, err := p.NewInteractionLogin(intent, clientID)
	if err != nil {
		return nil, err
	}
	i.UserID = userID
	ir := identity.ToRef()
	i.Identity = &ir
	if primaryAuthenticatorRef != nil {
		primaryAuthenticator, err := p.Authenticator.Get(
			userID,
			primaryAuthenticatorRef.Type,
			primaryAuthenticatorRef.ID)
		if err != nil {
			return nil, err
		}
		ar := primaryAuthenticator.ToRef()
		i.PrimaryAuthenticator = &ar
	}
	return i, nil
}

func (p *Provider) NewInteractionLogin(intent *IntentLogin, clientID string) (*Interaction, error) {
	switch intent.Identity.Type {
	case authn.IdentityTypeLoginID:
		return p.newInteractionLoginIDLogin(intent, clientID)
	case authn.IdentityTypeOAuth:
		return p.newInteractionOAuthLogin(intent, clientID)
	default:
		panic("interaction_provider: unknown identity type " + intent.Identity.Type)
	}
}

func (p *Provider) newInteractionLoginIDLogin(intent *IntentLogin, clientID string) (*Interaction, error) {
	i := &Interaction{
		Intent:   intent,
		ClientID: clientID,
	}
	return i, nil
}

func (p *Provider) newInteractionOAuthLogin(intent *IntentLogin, clientID string) (*Interaction, error) {
	i := &Interaction{
		Intent:               intent,
		ClientID:             clientID,
		PrimaryAuthenticator: nil,
		State:                map[string]string{},
	}
	userid, iden, err := p.Identity.GetByClaims(intent.Identity.Type, intent.Identity.Claims)
	if errors.Is(err, identity.ErrIdentityNotFound) {
		return nil, ErrInvalidCredentials
	} else if err != nil {
		return nil, err
	}
	i.UserID = userid
	ir := iden.ToRef()
	i.Identity = &ir
	return i, nil
}

func (p *Provider) NewInteractionSignup(intent *IntentSignup, clientID string) (*Interaction, error) {
	i := &Interaction{
		Intent:   intent,
		ClientID: clientID,
		UserID:   uuid.New(),
	}
	identity := p.Identity.New(i.UserID, intent.Identity.Type, intent.Identity.Claims)
	ir := identity.ToRef()
	i.Identity = &ir
	i.NewIdentities = append(i.NewIdentities, identity)

	if err := p.Identity.Validate(i.NewIdentities); err != nil {
		return nil, err
	}
	return i, nil
}

func (p *Provider) NewInteractionAddIdentity(intent *IntentAddIdentity, clientID string, userID string) (*Interaction, error) {
	i := &Interaction{
		Intent:   intent,
		ClientID: clientID,
		UserID:   userID,
	}
	identity := p.Identity.New(i.UserID, intent.Identity.Type, intent.Identity.Claims)
	ir := identity.ToRef()
	i.Identity = &ir
	i.NewIdentities = append(i.NewIdentities, identity)

	if err := p.Identity.Validate(i.NewIdentities); err != nil {
		return nil, err
	}
	if err := p.Identity.ValidateWithUser(i.UserID, i.NewIdentities); err != nil {
		return nil, err
	}
	return i, nil
}

func (p *Provider) NewInteractionAddAuthenticator(intent *IntentAddAuthenticator, clientID string, session auth.AuthSession) (*Interaction, error) {
	panic("TODO(interaction): implement it")
}
