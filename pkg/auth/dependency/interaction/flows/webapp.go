package flows

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity/loginid"
	"github.com/authgear/authgear-server/pkg/auth/dependency/interaction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

type WebAppFlow struct {
	Config         *config.IdentityConfig
	Identities     IdentityProvider
	Users          UserProvider
	Hooks          HookProvider
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

	step, err := f.handleLogin(i)
	if err != nil {
		return nil, err
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
	}, "")
	if err != nil {
		return nil, err
	}

	step, err := f.handleSignup(i)
	if err != nil {
		return nil, err
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

func (f *WebAppFlow) handleLogin(i *interaction.Interaction) (WebAppStep, error) {
	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return "", err
	}

	if s.CurrentStep().Step != interaction.StepAuthenticatePrimary || len(s.CurrentStep().AvailableAuthenticators) <= 0 {
		panic("interaction_flow_webapp: unexpected interaction state")
	}

	var step WebAppStep
	switch s.CurrentStep().AvailableAuthenticators[0].Type {
	case authn.AuthenticatorTypeOOB:
		step = WebAppStepAuthenticateOOBOTP
		err = f.Interactions.PerformAction(i, interaction.StepAuthenticatePrimary, &interaction.ActionTriggerOOBAuthenticator{
			Authenticator: s.CurrentStep().AvailableAuthenticators[0],
		})
		if err != nil {
			return "", err
		}
	case authn.AuthenticatorTypePassword:
		step = WebAppStepAuthenticatePassword
	default:
		panic("interaction_flow_webapp: unexpected authenticator type")
	}

	return step, nil
}

func (f *WebAppFlow) handleSignup(i *interaction.Interaction) (WebAppStep, error) {
	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return "", err
	}

	if s.CurrentStep().Step != interaction.StepSetupPrimaryAuthenticator || len(s.CurrentStep().AvailableAuthenticators) <= 0 {
		panic("interaction_flow_webapp: unexpected interaction state")
	}

	var step WebAppStep

	switch s.CurrentStep().AvailableAuthenticators[0].Type {
	case authn.AuthenticatorTypeOOB:
		step = WebAppStepSetupOOBOTP
		err = f.Interactions.PerformAction(i, interaction.StepSetupPrimaryAuthenticator, &interaction.ActionTriggerOOBAuthenticator{
			Authenticator: s.CurrentStep().AvailableAuthenticators[0],
		})
		if err != nil {
			return "", err
		}
	case authn.AuthenticatorTypePassword:
		step = WebAppStepSetupPassword
	default:
		panic("interaction_flow_webapp: unexpected authenticator type")
	}

	return step, nil
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

	if len(s.CurrentStep().AvailableAuthenticators) <= 0 {
		panic("interaction_flow_webapp: unexpected interaction state")
	}

	if s.CurrentStep().Step == interaction.StepSetupPrimaryAuthenticator {
		return f.SetupSecret(token, secret)
	}
	if s.CurrentStep().Step == interaction.StepAuthenticatePrimary {
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

	if len(s.CurrentStep().AvailableAuthenticators) <= 0 {
		panic("interaction_flow_webapp: unexpected interaction state")
	}

	err = f.Interactions.PerformAction(i, interaction.StepSetupPrimaryAuthenticator, &interaction.ActionSetupAuthenticator{
		Authenticator: s.CurrentStep().AvailableAuthenticators[0],
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

	result, err := f.Interactions.Commit(i)
	if err != nil {
		return nil, err
	}

	switch i.Intent.Type() {
	case interaction.IntentTypeSignup:
		// New interaction for logging in after signup
		i, err = f.Interactions.NewInteractionLoginAs(
			&interaction.IntentLogin{
				Identity: identity.Spec{
					Type:   result.Identity.Type,
					Claims: result.Identity.Claims,
				},
				OriginalIntentType: i.Intent.Type(),
			},
			result.Attrs.UserID,
			i.Identity,
			i.PrimaryAuthenticator,
			i.ClientID,
		)
		if err != nil {
			return nil, err
		}

		// Primary authentication is done using `AuthenticatedAs`
		return f.afterPrimaryAuthentication(i)

	case interaction.IntentTypeAddIdentity:
		if i.Extra[WebAppExtraStateAnonymousUserPromotion] != "" {
			return f.afterAnonymousUserPromotion(i, result)
		}

		return &WebAppResult{
			Step: WebAppStepCompleted,
		}, nil

	default:
		return &WebAppResult{
			Step: WebAppStepCompleted,
		}, nil
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

	if len(s.CurrentStep().AvailableAuthenticators) <= 0 {
		panic("interaction_flow_webapp: unexpected interaction state")
	}

	err = f.Interactions.PerformAction(i, interaction.StepAuthenticatePrimary, &interaction.ActionAuthenticate{
		Authenticator: s.CurrentStep().AvailableAuthenticators[0],
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

	if len(s.CurrentStep().AvailableAuthenticators) <= 0 || s.CurrentStep().AvailableAuthenticators[0].Type != authn.AuthenticatorTypeOOB {
		panic("interaction_flow_webapp: unexpected interaction state")
	}

	err = f.Interactions.PerformAction(i, s.CurrentStep().Step, &interaction.ActionTriggerOOBAuthenticator{
		Authenticator: s.CurrentStep().AvailableAuthenticators[0],
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

func (f *WebAppFlow) AddLoginID(userID string, loginID loginid.LoginID) (result *WebAppResult, err error) {
	clientID := ""
	i, err := f.Interactions.NewInteractionAddIdentity(&interaction.IntentAddIdentity{
		Identity: identity.Spec{
			Type: authn.IdentityTypeLoginID,
			Claims: map[string]interface{}{
				identity.IdentityClaimLoginIDKey:   loginID.Key,
				identity.IdentityClaimLoginIDValue: loginID.Value,
			},
		},
	}, clientID, userID)
	if err != nil {
		return
	}

	return f.afterAddUpdateRemoveLoginID(i)
}

func (f *WebAppFlow) RemoveLoginID(userID string, loginID loginid.LoginID) (result *WebAppResult, err error) {
	clientID := ""
	i, err := f.Interactions.NewInteractionRemoveIdentity(&interaction.IntentRemoveIdentity{
		Identity: identity.Spec{
			Type: authn.IdentityTypeLoginID,
			Claims: map[string]interface{}{
				identity.IdentityClaimLoginIDKey:   loginID.Key,
				identity.IdentityClaimLoginIDValue: loginID.Value,
			},
		},
	}, clientID, userID)
	if err != nil {
		return
	}

	return f.afterAddUpdateRemoveLoginID(i)
}

func (f *WebAppFlow) UpdateLoginID(userID string, oldLoginID loginid.LoginID, newLoginID loginid.LoginID) (result *WebAppResult, err error) {
	clientID := ""
	i, err := f.Interactions.NewInteractionUpdateIdentity(&interaction.IntentUpdateIdentity{
		OldIdentity: identity.Spec{
			Type: authn.IdentityTypeLoginID,
			Claims: map[string]interface{}{
				identity.IdentityClaimLoginIDKey:   oldLoginID.Key,
				identity.IdentityClaimLoginIDValue: oldLoginID.Value,
			},
		},
		NewIdentity: identity.Spec{
			Type: authn.IdentityTypeLoginID,
			Claims: map[string]interface{}{
				identity.IdentityClaimLoginIDKey:   newLoginID.Key,
				identity.IdentityClaimLoginIDValue: newLoginID.Value,
			},
		},
	}, clientID, userID)
	if err != nil {
		return nil, err
	}

	return f.afterAddUpdateRemoveLoginID(i)
}

func (f *WebAppFlow) afterAddUpdateRemoveLoginID(i *interaction.Interaction) (result *WebAppResult, err error) {
	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return nil, err
	}

	// Either commit
	if s.CurrentStep().Step == interaction.StepCommit {
		_, err = f.Interactions.Commit(i)
		if err != nil {
			return
		}

		result = &WebAppResult{
			Step: WebAppStepCompleted,
		}

		return
	}

	// Or have more steps to go through
	if s.CurrentStep().Step != interaction.StepSetupPrimaryAuthenticator || len(s.CurrentStep().AvailableAuthenticators) <= 0 {
		panic("interaction_flow_webapp: unexpected interaction state")
	}

	var step WebAppStep
	switch s.CurrentStep().AvailableAuthenticators[0].Type {
	case authn.AuthenticatorTypeOOB:
		step = WebAppStepSetupOOBOTP
		err = f.Interactions.PerformAction(i, interaction.StepSetupPrimaryAuthenticator, &interaction.ActionTriggerOOBAuthenticator{
			Authenticator: s.CurrentStep().AvailableAuthenticators[0],
		})
		if err != nil {
			return nil, err
		}
	case authn.AuthenticatorTypePassword:
		step = WebAppStepSetupPassword
	default:
		panic("interaction_flow_webapp: unexpected authenticator type")
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

func (f *WebAppFlow) afterPrimaryAuthentication(i *interaction.Interaction) (*WebAppResult, error) {
	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return nil, err
	}
	switch s.CurrentStep().Step {
	case interaction.StepAuthenticateSecondary, interaction.StepSetupSecondaryAuthenticator:
		panic("interaction_flow_webapp: TODO: handle MFA")

	case interaction.StepCommit:
		ir, err := f.Interactions.Commit(i)
		if err != nil {
			return nil, err
		}

		if i.Extra[WebAppExtraStateAnonymousUserPromotion] != "" {
			return f.afterAnonymousUserPromotion(i, ir)
		}

		result, err := f.UserController.CreateSession(i, ir)
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
