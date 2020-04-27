package interaction

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

func (p *Provider) NewInteractionLogin(intent *IntentLogin, clientID string) (*Interaction, error) {
	i := &Interaction{
		Intent:   intent,
		ClientID: clientID,
	}
	if intent.AuthenticatedAs != nil {
		identity, err := p.Identity.Get(
			intent.AuthenticatedAs.UserID,
			intent.Identity.Type,
			intent.Identity.ID)
		if err != nil {
			return nil, err
		}
		authenticator, err := p.Authenticator.Get(
			intent.AuthenticatedAs.UserID,
			intent.AuthenticatedAs.PrimaryAuthenticator.Type,
			intent.AuthenticatedAs.PrimaryAuthenticator.ID)
		if err != nil {
			return nil, err
		}
		i.UserID = intent.AuthenticatedAs.UserID
		i.Identity = identity
		i.PrimaryAuthenticator = authenticator
	}
	return i, nil
}

func (p *Provider) NewInteractionSignup(intent *IntentSignup, clientID string) (*Interaction, error) {
	i := &Interaction{
		Intent:   intent,
		ClientID: clientID,
		UserID:   uuid.New(),
	}
	identity := p.Identity.New(i.UserID, intent.Identity.Type, intent.Identity.Claims)
	i.Identity = identity
	i.NewIdentities = append(i.NewIdentities, identity)

	if err := p.Identity.Validate(i.NewIdentities); err != nil {
		return nil, err
	}
	return i, nil
}

func (p *Provider) NewInteractionAddAuthenticator(intent *IntentAddAuthenticator, clientID string, session auth.AuthSession) (*Interaction, error) {
	panic("TODO(interaction): implement it")
}
