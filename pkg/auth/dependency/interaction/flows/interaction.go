package flows

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/interaction"
)

type InteractionProvider interface {
	Commit(*interaction.Interaction) (*interaction.Result, error)
	NewInteractionLogin(intent *interaction.IntentLogin, clientID string) (*interaction.Interaction, error)
	NewInteractionLoginAs(
		intent *interaction.IntentLogin,
		userID string,
		identityRef *identity.Ref,
		primaryAuthenticatorRef *authenticator.Ref,
		clientID string,
	) (*interaction.Interaction, error)
	NewInteractionSignup(intent *interaction.IntentSignup, clientID string) (*interaction.Interaction, error)
	NewInteractionAddIdentity(intent *interaction.IntentAddIdentity, clientID string, userID string) (*interaction.Interaction, error)
	NewInteractionRemoveIdentity(intent *interaction.IntentRemoveIdentity, clientID string, userID string) (*interaction.Interaction, error)
	NewInteractionUpdateIdentity(intent *interaction.IntentUpdateIdentity, clientID string, userID string) (*interaction.Interaction, error)
	NewInteractionUpdateAuthenticator(intent *interaction.IntentUpdateAuthenticator, clientID string, userID string) (*interaction.Interaction, error)
	GetInteractionState(i *interaction.Interaction) (*interaction.State, error)
	PerformAction(i *interaction.Interaction, step interaction.Step, action interaction.Action) error
}
