package flows

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

type InteractionProvider interface {
	GetInteraction(token string) (*interaction.Interaction, error)
	SaveInteraction(*interaction.Interaction) (string, error)
	Commit(*interaction.Interaction) (*authn.Attrs, error)
	NewInteractionLogin(intent *interaction.IntentLogin, clientID string) (*interaction.Interaction, error)
	GetInteractionState(i *interaction.Interaction) (*interaction.State, error)
	PerformAction(i *interaction.Interaction, step interaction.Step, action interaction.Action) error
}
