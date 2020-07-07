package flows

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/interaction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

type PasswordFlow struct {
	Interactions InteractionProvider
}

func (f *PasswordFlow) ChangePassword(userID string, OldPassword string, newPassword string) error {
	i, err := f.Interactions.NewInteractionUpdateAuthenticator(&interaction.IntentUpdateAuthenticator{
		Authenticator: authenticator.Spec{
			Type: authn.AuthenticatorTypePassword,
		},
		OldSecret: OldPassword,
	}, "", userID)

	if err != nil {
		return err
	}

	return f.startUpdatePasswordInteraction(i, newPassword)
}

func (f *PasswordFlow) ResetPassword(userID string, password string) error {
	i, err := f.Interactions.NewInteractionUpdateAuthenticator(&interaction.IntentUpdateAuthenticator{
		Authenticator: authenticator.Spec{
			Type: authn.AuthenticatorTypePassword,
		},
		SkipVerifySecret: true,
	}, "", userID)

	if err != nil {
		return err
	}

	return f.startUpdatePasswordInteraction(i, password)
}

func (f *PasswordFlow) startUpdatePasswordInteraction(i *interaction.Interaction, password string) error {
	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return err
	}

	if s.CurrentStep().Step != interaction.StepSetupPrimaryAuthenticator ||
		len(s.CurrentStep().AvailableAuthenticators) != 1 ||
		s.CurrentStep().AvailableAuthenticators[0].Type != authn.AuthenticatorTypePassword {
		panic("interaction_flow_password: unexpected interaction state")
	}

	passwordAuthenticator := s.CurrentStep().AvailableAuthenticators[0]
	err = f.Interactions.PerformAction(i, interaction.StepSetupPrimaryAuthenticator, &interaction.ActionSetupAuthenticator{
		Authenticator: passwordAuthenticator,
		Secret:        password,
	})
	if err != nil {
		return err
	}

	s, err = f.Interactions.GetInteractionState(i)
	if err != nil {
		return err
	}

	if s.CurrentStep().Step != interaction.StepCommit {
		panic("interaction_flow_password: unexpected interaction state")
	}

	_, err = f.Interactions.Commit(i)

	return err
}
