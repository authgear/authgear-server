package interaction

import (
	"errors"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

func (p *Provider) NewInteractionLoginAs(
	intent *IntentLogin,
	userID string,
	identityRef *identity.Ref,
	primaryAuthenticatorRef *authenticator.Ref,
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
		return p.newInteractionLogin(intent, clientID)
	case authn.IdentityTypeOAuth, authn.IdentityTypeAnonymous:
		return p.newInteractionLoginRequireIdentity(intent, clientID)
	default:
		panic("interaction_provider: unknown identity type " + intent.Identity.Type)
	}
}

func (p *Provider) newInteractionLogin(intent *IntentLogin, clientID string) (*Interaction, error) {
	i := newInteraction(clientID, intent)
	return i, nil
}

func (p *Provider) newInteractionLoginRequireIdentity(intent *IntentLogin, clientID string) (*Interaction, error) {
	i := newInteraction(clientID, intent)

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
	i := newInteraction(clientID, intent)
	i.UserID = uuid.New()

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
	i := newInteraction(clientID, intent)
	i.UserID = userID

	id := p.Identity.New(i.UserID, intent.Identity.Type, intent.Identity.Claims)
	ir := id.ToRef()
	i.Identity = &ir
	i.NewIdentities = append(i.NewIdentities, id)

	ois, err := p.Identity.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	existingIdentities := []*identity.Info{}
	for _, oi := range ois {
		if oi.Type == id.Type {
			existingIdentities = append(existingIdentities, oi)
		}
	}
	checkIdentities := append(i.NewIdentities, existingIdentities...)

	if err := p.Identity.Validate(checkIdentities); err != nil {
		if errors.Is(err, identity.ErrIdentityAlreadyExists) {
			return nil, ErrDuplicatedIdentity
		}
		return nil, err
	}
	return i, nil
}

func (p *Provider) NewInteractionUpdateIdentity(intent *IntentUpdateIdentity, clientID string, userID string) (*Interaction, error) {
	i := newInteraction(clientID, intent)
	i.UserID = userID

	if intent.OldIdentity.Type != intent.NewIdentity.Type {
		panic("interaction: update identity type is not expected")
	}

	uid, oldIden, err := p.Identity.GetByClaims(intent.OldIdentity.Type, intent.OldIdentity.Claims)
	if errors.Is(err, identity.ErrIdentityNotFound) || uid != userID {
		return nil, ErrIdentityNotFound
	} else if err != nil {
		return nil, err
	}

	updateIden := p.Identity.WithClaims(userID, oldIden, intent.NewIdentity.Claims)
	ir := oldIden.ToRef()
	i.Identity = &ir
	i.UpdateIdentities = append(i.UpdateIdentities, updateIden)

	ois, err := p.Identity.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	checkIdentities := []*identity.Info{}
	for _, oi := range ois {
		if oi.Type == updateIden.Type {
			if oi.ID == updateIden.ID {
				checkIdentities = append(checkIdentities, updateIden)
			} else {
				checkIdentities = append(checkIdentities, oi)
			}
		}
	}

	if err := p.Identity.Validate(checkIdentities); err != nil {
		return nil, err
	}
	return i, nil
}

func (p *Provider) NewInteractionRemoveIdentity(intent *IntentRemoveIdentity, clientID string, userID string) (*Interaction, error) {
	i := newInteraction(clientID, intent)
	i.UserID = userID

	iden, err := p.Identity.GetByUserAndClaims(intent.Identity.Type, userID, intent.Identity.Claims)
	if errors.Is(err, identity.ErrIdentityNotFound) {
		return nil, ErrIdentityNotFound
	} else if err != nil {
		return nil, err
	}

	ois, err := p.Identity.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	if len(ois) <= 1 {
		return nil, ErrCannotRemoveLastIdentity
	}

	ir := iden.ToRef()
	i.Identity = &ir
	i.RemoveIdentities = append(i.RemoveIdentities, iden)
	return i, nil
}

func (p *Provider) NewInteractionAddAuthenticator(intent *IntentAddAuthenticator, clientID string, session auth.AuthSession) (*Interaction, error) {
	panic("TODO(interaction): implement it")
}

func (p *Provider) NewInteractionUpdateAuthenticator(
	intent *IntentUpdateAuthenticator, clientID string, userID string,
) (*Interaction, error) {
	if intent.Authenticator.Type != authn.AuthenticatorTypePassword {
		panic("interaction: update authenticator is not supported for type " + intent.Authenticator.Type)
	}
	i := &Interaction{
		Intent:   intent,
		ClientID: clientID,
		UserID:   userID,
	}
	ais, err := p.Authenticator.List(i.UserID, intent.Authenticator.Type)
	if err != nil {
		return nil, err
	}
	if len(ais) != 1 {
		return nil, ErrAuthenticatorNotFound
	}
	authen := ais[0]
	if !intent.SkipVerifySecret {
		err = p.Authenticator.VerifySecret(userID, authen, intent.OldSecret)
		if err != nil {
			return nil, err
		}
	}

	i.PendingAuthenticator = authen
	return i, nil
}
