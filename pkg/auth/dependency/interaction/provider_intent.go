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
	return i, nil
}

func (p *Provider) NewInteractionAddAuthenticator(intent *IntentAddAuthenticator, clientID string, session auth.AuthSession) (*Interaction, error) {
	panic("TODO(interaction): implement it")
}
