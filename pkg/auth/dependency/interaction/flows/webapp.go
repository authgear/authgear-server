package flows

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

type WebAppFlow struct {
	Interactions   InteractionProvider
	UserController *UserController
}

func (f *WebAppFlow) LoginWithLoginID(loginID string) (*TokenResult, error) {
	i, err := f.Interactions.NewInteraction(&interaction.IntentLogin{
		Identity: interaction.IdentitySpec{
			Type: authn.IdentityTypeLoginID,
			Claims: map[string]interface{}{
				interaction.IdentityClaimLoginIDValue: loginID,
			},
		},
	}, "", nil)
	if err != nil {
		return nil, err
	}

	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return nil, err
	} else if len(s.Steps) != 1 || s.Steps[0].Step != interaction.StepAuthenticatePrimary {
		panic("interaction_flow_webapp: unexpected interaction state")
	}

	token, err := f.Interactions.SaveInteraction(i)
	if err != nil {
		return nil, err
	}

	return &TokenResult{
		Token: token,
	}, nil
}

func (f *WebAppFlow) AuthenticatePassword(token string, password string) (*WebAppResult, error) {
	i, err := f.Interactions.GetInteraction(token)
	if err != nil {
		return nil, err
	}

	err = f.Interactions.PerformAction(i, interaction.StepAuthenticatePrimary, &interaction.ActionAuthenticate{
		Authenticator: interaction.AuthenticatorSpec{Type: interaction.AuthenticatorTypePassword},
		Secret:        password,
	})
	if err != nil {
		return nil, err
	}

	_, err = f.Interactions.SaveInteraction(i)
	if err != nil {
		return nil, err
	}

	if i.Error != nil {
		return nil, i.Error
	}

	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return nil, err
	}

	switch s.CurrentStep().Step {
	case interaction.StepAuthenticateSecondary, interaction.StepSetupSecondaryAuthenticator:
		panic("interaction_flow_webapp: TODO: handle MFA")

	case interaction.StepCommit:
		attrs, err := f.Interactions.Commit(i)
		if err != nil {
			return nil, err
		}

		result, err := f.UserController.CreateSession(i, attrs, false)
		if err != nil {
			return nil, err
		}

		return &WebAppResult{Cookies: result.Cookies}, nil

	default:
		panic("interaction_flow_webapp: unexpected step " + s.CurrentStep().Step)
	}
}
