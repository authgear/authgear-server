package flows

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

type WebAppFlow struct {
	Interactions   InteractionProvider
	UserController *UserController
}

func (f *WebAppFlow) LoginWithLoginID(loginID string) (*WebAppResult, error) {
	i, err := f.Interactions.NewInteractionLogin(&interaction.IntentLogin{
		Identity: identity.Spec{
			Type: authn.IdentityTypeLoginID,
			Claims: map[string]interface{}{
				identity.IdentityClaimLoginIDValue: loginID,
			},
		},
	}, "")
	if err != nil {
		return nil, err
	}

	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return nil, err
	} else if len(s.Steps) != 1 || s.Steps[0].Step != interaction.StepAuthenticatePrimary {
		panic("interaction_flow_webapp: unexpected interaction state")
	}

	step := WebAppStepAuthenticatePassword
	// Trigger OOB OTP if the first authenticator is OOB OTP
	if len(s.Steps[0].AvailableAuthenticators) > 0 && s.Steps[0].AvailableAuthenticators[0].Type == authn.AuthenticatorTypeOOB {
		step = WebAppStepAuthenticateOOBOTP
		err = f.Interactions.PerformAction(i, interaction.StepAuthenticatePrimary, &interaction.ActionTriggerOOBAuthenticator{
			Authenticator: s.Steps[0].AvailableAuthenticators[0],
		})
		if err != nil {
			return nil, err
		}
	}

	token, err := f.Interactions.SaveInteraction(i)
	if err != nil {
		return nil, err
	}

	return &WebAppResult{
		Step:  step,
		Token: token,
	}, nil
}

func (f *WebAppFlow) SignupWithLoginID(loginIDKey, loginID string) (*WebAppResult, error) {
	i, err := f.Interactions.NewInteractionSignup(&interaction.IntentSignup{
		Identity: identity.Spec{
			Type: authn.IdentityTypeLoginID,
			Claims: map[string]interface{}{
				identity.IdentityClaimLoginIDKey:   loginIDKey,
				identity.IdentityClaimLoginIDValue: loginID,
			},
		},
		OnUserDuplicate: model.OnUserDuplicateAbort,
	}, "")
	if err != nil {
		return nil, err
	}

	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return nil, err
	} else if len(s.Steps) != 1 || s.Steps[0].Step != interaction.StepSetupPrimaryAuthenticator {
		panic("interaction_flow_webapp: unexpected interaction state")
	}

	step := WebAppStepSetupPassword
	// Trigger OOB OTP if the first authenticator is OOB OTP
	if len(s.Steps[0].AvailableAuthenticators) > 0 && s.Steps[0].AvailableAuthenticators[0].Type == authn.AuthenticatorTypeOOB {
		step = WebAppStepSetupOOBOTP
		err = f.Interactions.PerformAction(i, interaction.StepSetupPrimaryAuthenticator, &interaction.ActionTriggerOOBAuthenticator{
			Authenticator: s.Steps[0].AvailableAuthenticators[0],
		})
		if err != nil {
			return nil, err
		}
	}

	token, err := f.Interactions.SaveInteraction(i)
	if err != nil {
		return nil, err
	}

	return &WebAppResult{
		Step:  step,
		Token: token,
	}, nil
}

func (f *WebAppFlow) EnterSecret(token string, secret string) (*WebAppResult, error) {
	i, err := f.Interactions.GetInteraction(token)
	if err != nil {
		return nil, err
	}

	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return nil, err
	}

	if len(s.Steps) <= 0 || len(s.Steps[0].AvailableAuthenticators) <= 0 {
		panic("interaction_flow_webapp: unexpected interaction state")
	}

	if s.Steps[0].Step == interaction.StepSetupPrimaryAuthenticator {
		return f.SetupSecret(token, secret)
	}
	if s.Steps[0].Step == interaction.StepAuthenticatePrimary {
		return f.AuthenticateSecret(token, secret)
	}

	panic("interaction_flow_webapp: unexpected interaction state")
}

func (f *WebAppFlow) SetupSecret(token string, secret string) (*WebAppResult, error) {
	i, err := f.Interactions.GetInteraction(token)
	if err != nil {
		return nil, err
	}

	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return nil, err
	}

	if len(s.Steps) <= 0 || len(s.Steps[0].AvailableAuthenticators) <= 0 {
		panic("interaction_flow_webapp: unexpected interaction state")
	}

	err = f.Interactions.PerformAction(i, interaction.StepSetupPrimaryAuthenticator, &interaction.ActionSetupAuthenticator{
		Authenticator: s.Steps[0].AvailableAuthenticators[0],
		Secret:        secret,
	})
	if err != nil {
		return nil, err
	}

	if i.Error != nil {
		return nil, i.Error
	}

	s, err = f.Interactions.GetInteractionState(i)
	if err != nil {
		return nil, err
	} else if s.CurrentStep().Step != interaction.StepCommit {
		panic("interaction_flow_webapp: unexpected interaction state")
	}

	attrs, err := f.Interactions.Commit(i)
	if err != nil {
		return nil, err
	}

	switch i.Intent.Type() {
	case interaction.IntentTypeSignup:
		// New interaction for logging in after signup
		i, err = f.Interactions.NewInteractionLoginAs(
			&interaction.IntentLogin{
				Identity: identity.Spec{
					Type:   attrs.IdentityType,
					Claims: attrs.IdentityClaims,
				},
				OriginalIntentType: i.Intent.Type(),
			},
			attrs.UserID,
			i.Identity,
			i.PrimaryAuthenticator,
			i.ClientID,
		)
		if err != nil {
			return nil, err
		}

		// Primary authentication is done using `AuthenticatedAs`
		return f.afterPrimaryAuthentication(i)

	default:
		panic("interaction_flow_webapp: unexpected interaction intent")
	}
}

func (f *WebAppFlow) AuthenticateSecret(token string, secret string) (*WebAppResult, error) {
	i, err := f.Interactions.GetInteraction(token)
	if err != nil {
		return nil, err
	}

	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return nil, err
	}

	if len(s.Steps) <= 0 || len(s.Steps[0].AvailableAuthenticators) <= 0 {
		panic("interaction_flow_webapp: unexpected interaction state")
	}

	err = f.Interactions.PerformAction(i, interaction.StepAuthenticatePrimary, &interaction.ActionAuthenticate{
		Authenticator: s.Steps[0].AvailableAuthenticators[0],
		Secret:        secret,
	})
	if err != nil {
		return nil, err
	}

	if i.Error != nil {
		return nil, i.Error
	}

	return f.afterPrimaryAuthentication(i)
}

func (f *WebAppFlow) TriggerOOBOTP(token string) (*WebAppResult, error) {
	i, err := f.Interactions.GetInteraction(token)
	if err != nil {
		return nil, err
	}

	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return nil, err
	}

	if len(s.Steps) <= 0 || len(s.Steps[0].AvailableAuthenticators) <= 0 || s.Steps[0].AvailableAuthenticators[0].Type != authn.AuthenticatorTypeOOB {
		panic("interaction_flow_webapp: unexpected interaction state")
	}

	err = f.Interactions.PerformAction(i, s.Steps[0].Step, &interaction.ActionTriggerOOBAuthenticator{
		Authenticator: s.Steps[0].AvailableAuthenticators[0],
	})
	if err != nil {
		return nil, err
	}

	token, err = f.Interactions.SaveInteraction(i)
	if err != nil {
		return nil, err
	}

	return &WebAppResult{
		Step:  WebAppStepAuthenticateOOBOTP,
		Token: token,
	}, nil
}

func (f *WebAppFlow) afterPrimaryAuthentication(i *interaction.Interaction) (*WebAppResult, error) {
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

		return &WebAppResult{
			Step:    WebAppStepCompleted,
			Cookies: result.Cookies,
		}, nil

	default:
		panic("interaction_flow_webapp: unexpected step " + s.CurrentStep().Step)
	}
}
