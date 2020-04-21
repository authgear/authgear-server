package flows

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

type AuthAPIFlow struct {
	Interactions   InteractionProvider
	UserController *UserController
}

func (f *AuthAPIFlow) LoginWithLoginIDPassword(
	clientID string, loginIDKey string, loginID string, password string,
) (*AuthResult, error) {
	i, err := f.Interactions.NewInteraction(&interaction.IntentLogin{
		Identity: interaction.IdentitySpec{
			Type: authn.IdentityTypeLoginID,
			Claims: map[string]interface{}{
				interaction.IdentityClaimLoginIDKey:   loginIDKey,
				interaction.IdentityClaimLoginIDValue: loginID,
			},
		},
	}, clientID, nil)
	if err != nil {
		return nil, err
	}

	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return nil, err
	} else if s.CurrentStep().Step != interaction.StepAuthenticatePrimary {
		return nil, ErrUnsupportedConfiguration
	}

	err = f.Interactions.PerformAction(
		i,
		interaction.StepAuthenticatePrimary,
		&interaction.ActionAuthenticate{
			Authenticator: interaction.AuthenticatorSpec{Type: interaction.AuthenticatorTypePassword},
			Secret:        password,
		},
	)
	if err != nil {
		return nil, err
	}

	s, err = f.Interactions.GetInteractionState(i)
	if err != nil {
		return nil, err
	}

	switch s.CurrentStep().Step {
	case interaction.StepAuthenticateSecondary, interaction.StepSetupSecondaryAuthenticator:
		panic("interaction_flow_auth_api: TODO: handle MFA")

	case interaction.StepCommit:
		attrs, err := f.Interactions.Commit(i)
		if err != nil {
			return nil, err
		}

		result, err := f.UserController.CreateSession(i, attrs, true)
		if err != nil {
			return nil, err
		}

		return result, nil

	default:
		return nil, ErrUnsupportedConfiguration
	}
}
