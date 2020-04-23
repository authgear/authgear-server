package interaction

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
)

func (p *Provider) NewInteractionLogin(intent *IntentLogin, clientID string) (*Interaction, error) {
	i := &Interaction{
		Intent:   intent,
		ClientID: clientID,
	}
	return i, nil
}

func (p *Provider) NewInteractionSignup(intent *IntentSignup, clientID string) (*Interaction, error) {
	panic("TODO(interaction): implement it")
}

func (p *Provider) NewInteractionAddAuthenticator(intent *IntentAddAuthenticator, clientID string, session auth.AuthSession) (*Interaction, error) {
	panic("TODO(interaction): implement it")
}
