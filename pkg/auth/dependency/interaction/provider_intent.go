package interaction

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
)

func (p *Provider) NewInteraction(intent Intent, clientID string, session auth.AuthSession) (*Interaction, error) {
	switch intent := intent.(type) {
	case *IntentLogin:
		return p.newInteractionLogin(intent, clientID)
	}
	panic(fmt.Sprintf("interaction: unknown intent type %T", intent))
}

func (p *Provider) newInteractionLogin(intent *IntentLogin, clientID string) (*Interaction, error) {
	i := &Interaction{
		Intent:   intent,
		ClientID: clientID,
	}
	return i, nil
}
