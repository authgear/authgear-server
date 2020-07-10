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

func (f *WebAppFlow) GetInteractionState(i *interaction.Interaction) (*interaction.State, error) {
	return f.Interactions.GetInteractionState(i)
}

func (f *WebAppFlow) LoginWithLoginID(state *State, loginID string) (*WebAppResult, error) {
	var err error
	state.Interaction, err = f.Interactions.NewInteractionLogin(&interaction.IntentLogin{
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

	err = f.handleLogin(state)
	if err != nil {
		return nil, err
	}

	state.Extra[ExtraGivenLoginID] = loginID

	return &WebAppResult{}, nil
}

func (f *WebAppFlow) SignupWithLoginID(state *State, loginIDKey, loginID string) (*WebAppResult, error) {
	var err error
	state.Interaction, err = f.Interactions.NewInteractionSignup(&interaction.IntentSignup{
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

	err = f.handleSignup(state)
	if err != nil {
		return nil, err
	}

	state.Extra[ExtraGivenLoginID] = loginID

	return &WebAppResult{}, nil
}

func (f *WebAppFlow) handleLogin(state *State) error {
	s, err := f.Interactions.GetInteractionState(state.Interaction)
	if err != nil {
		return err
	}

	if s.CurrentStep().Step != interaction.StepAuthenticatePrimary || len(s.CurrentStep().AvailableAuthenticators) <= 0 {
		panic("interaction_flow_webapp: unexpected interaction state")
	}

	switch s.CurrentStep().AvailableAuthenticators[0].Type {
	case authn.AuthenticatorTypeOOB:
		err = f.Interactions.PerformAction(state.Interaction, interaction.StepAuthenticatePrimary, &interaction.ActionTriggerOOBAuthenticator{
			Authenticator: s.CurrentStep().AvailableAuthenticators[0],
		})
		if err != nil {
			return err
		}
	case authn.AuthenticatorTypePassword:
		break
	default:
		panic("interaction_flow_webapp: unexpected authenticator type")
	}

	return nil
}

func (f *WebAppFlow) handleSignup(state *State) error {
	s, err := f.Interactions.GetInteractionState(state.Interaction)
	if err != nil {
		return err
	}

	if s.CurrentStep().Step != interaction.StepSetupPrimaryAuthenticator || len(s.CurrentStep().AvailableAuthenticators) <= 0 {
		panic("interaction_flow_webapp: unexpected interaction state")
	}

	switch s.CurrentStep().AvailableAuthenticators[0].Type {
	case authn.AuthenticatorTypeOOB:
		err = f.Interactions.PerformAction(state.Interaction, interaction.StepSetupPrimaryAuthenticator, &interaction.ActionTriggerOOBAuthenticator{
			Authenticator: s.CurrentStep().AvailableAuthenticators[0],
		})
		if err != nil {
			return err
		}
	case authn.AuthenticatorTypePassword:
		break
	default:
		panic("interaction_flow_webapp: unexpected authenticator type")
	}

	return nil
}

func (f *WebAppFlow) EnterSecret(state *State, secret string) (*WebAppResult, error) {
	s, err := f.Interactions.GetInteractionState(state.Interaction)
	if err != nil {
		return nil, err
	}

	if len(s.CurrentStep().AvailableAuthenticators) <= 0 {
		panic("interaction_flow_webapp: unexpected interaction state")
	}

	if s.CurrentStep().Step == interaction.StepSetupPrimaryAuthenticator {
		return f.SetupSecret(state, secret)
	}
	if s.CurrentStep().Step == interaction.StepAuthenticatePrimary {
		return f.AuthenticateSecret(state, secret)
	}

	panic("interaction_flow_webapp: unexpected interaction state")
}

func (f *WebAppFlow) SetupSecret(state *State, secret string) (*WebAppResult, error) {
	s, err := f.Interactions.GetInteractionState(state.Interaction)
	if err != nil {
		return nil, err
	}

	if len(s.CurrentStep().AvailableAuthenticators) <= 0 {
		panic("interaction_flow_webapp: unexpected interaction state")
	}

	err = f.Interactions.PerformAction(state.Interaction, interaction.StepSetupPrimaryAuthenticator, &interaction.ActionSetupAuthenticator{
		Authenticator: s.CurrentStep().AvailableAuthenticators[0],
		Secret:        secret,
	})
	if err != nil {
		return nil, err
	}

	s, err = f.Interactions.GetInteractionState(state.Interaction)
	if err != nil {
		return nil, err
	} else if s.CurrentStep().Step != interaction.StepCommit {
		panic("interaction_flow_webapp: unexpected interaction state")
	}

	result, err := f.Interactions.Commit(state.Interaction)
	if err != nil {
		return nil, err
	}

	switch state.Interaction.Intent.Type() {
	case interaction.IntentTypeSignup:
		// New interaction for logging in after signup
		state.Interaction, err = f.Interactions.NewInteractionLoginAs(
			&interaction.IntentLogin{
				Identity: identity.Spec{
					Type:   result.Identity.Type,
					Claims: result.Identity.Claims,
				},
				OriginalIntentType: state.Interaction.Intent.Type(),
			},
			result.Attrs.UserID,
			state.Interaction.Identity,
			state.Interaction.PrimaryAuthenticator,
			state.Interaction.ClientID,
		)
		if err != nil {
			return nil, err
		}

		// Primary authentication is done using `AuthenticatedAs`
		return f.afterPrimaryAuthentication(state)
	case interaction.IntentTypeAddIdentity:
		if _, ok := state.Extra[WebAppExtraStateAnonymousUserPromotion].(string); ok {
			return f.afterAnonymousUserPromotion(state, result)
		}

		return &WebAppResult{}, nil

	default:
		return &WebAppResult{}, nil
	}
}

func (f *WebAppFlow) AuthenticateSecret(state *State, secret string) (*WebAppResult, error) {
	s, err := f.Interactions.GetInteractionState(state.Interaction)
	if err != nil {
		return nil, err
	}

	if len(s.CurrentStep().AvailableAuthenticators) <= 0 {
		panic("interaction_flow_webapp: unexpected interaction state")
	}

	err = f.Interactions.PerformAction(state.Interaction, interaction.StepAuthenticatePrimary, &interaction.ActionAuthenticate{
		Authenticator: s.CurrentStep().AvailableAuthenticators[0],
		Secret:        secret,
	})
	if err != nil {
		return nil, err
	}

	return f.afterPrimaryAuthentication(state)
}

func (f *WebAppFlow) TriggerOOBOTP(state *State) (*WebAppResult, error) {
	s, err := f.Interactions.GetInteractionState(state.Interaction)
	if err != nil {
		return nil, err
	}

	if len(s.CurrentStep().AvailableAuthenticators) <= 0 || s.CurrentStep().AvailableAuthenticators[0].Type != authn.AuthenticatorTypeOOB {
		panic("interaction_flow_webapp: unexpected interaction state")
	}

	err = f.Interactions.PerformAction(state.Interaction, s.CurrentStep().Step, &interaction.ActionTriggerOOBAuthenticator{
		Authenticator: s.CurrentStep().AvailableAuthenticators[0],
	})
	if err != nil {
		return nil, err
	}

	return &WebAppResult{}, nil
}

func (f *WebAppFlow) AddLoginID(state *State, userID string, loginID loginid.LoginID) (result *WebAppResult, err error) {
	clientID := ""
	state.Interaction, err = f.Interactions.NewInteractionAddIdentity(&interaction.IntentAddIdentity{
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

	state.Extra[ExtraGivenLoginID] = loginID.Value

	return f.afterAddUpdateRemoveLoginID(state)
}

func (f *WebAppFlow) RemoveLoginID(state *State, userID string, loginID loginid.LoginID) (result *WebAppResult, err error) {
	clientID := ""
	state.Interaction, err = f.Interactions.NewInteractionRemoveIdentity(&interaction.IntentRemoveIdentity{
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

	return f.afterAddUpdateRemoveLoginID(state)
}

func (f *WebAppFlow) UpdateLoginID(state *State, userID string, oldLoginID loginid.LoginID, newLoginID loginid.LoginID) (result *WebAppResult, err error) {
	clientID := ""
	state.Interaction, err = f.Interactions.NewInteractionUpdateIdentity(&interaction.IntentUpdateIdentity{
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

	return f.afterAddUpdateRemoveLoginID(state)
}

func (f *WebAppFlow) afterAddUpdateRemoveLoginID(state *State) (result *WebAppResult, err error) {
	s, err := f.Interactions.GetInteractionState(state.Interaction)
	if err != nil {
		return nil, err
	}

	// Either commit
	if s.CurrentStep().Step == interaction.StepCommit {
		_, err = f.Interactions.Commit(state.Interaction)
		if err != nil {
			return
		}

		result = &WebAppResult{}
		return
	}

	// Or have more steps to go through
	if s.CurrentStep().Step != interaction.StepSetupPrimaryAuthenticator || len(s.CurrentStep().AvailableAuthenticators) <= 0 {
		panic("interaction_flow_webapp: unexpected interaction state")
	}

	switch s.CurrentStep().AvailableAuthenticators[0].Type {
	case authn.AuthenticatorTypeOOB:
		err = f.Interactions.PerformAction(state.Interaction, interaction.StepSetupPrimaryAuthenticator, &interaction.ActionTriggerOOBAuthenticator{
			Authenticator: s.CurrentStep().AvailableAuthenticators[0],
		})
		if err != nil {
			return nil, err
		}
	case authn.AuthenticatorTypePassword:
		break
	default:
		panic("interaction_flow_webapp: unexpected authenticator type")
	}

	return &WebAppResult{}, nil
}

func (f *WebAppFlow) afterPrimaryAuthentication(state *State) (*WebAppResult, error) {
	s, err := f.Interactions.GetInteractionState(state.Interaction)
	if err != nil {
		return nil, err
	}
	switch s.CurrentStep().Step {
	case interaction.StepAuthenticateSecondary, interaction.StepSetupSecondaryAuthenticator:
		panic("interaction_flow_webapp: TODO: handle MFA")

	case interaction.StepCommit:
		ir, err := f.Interactions.Commit(state.Interaction)
		if err != nil {
			return nil, err
		}

		if _, ok := state.Extra[WebAppExtraStateAnonymousUserPromotion].(string); ok {
			return f.afterAnonymousUserPromotion(state, ir)
		}

		result, err := f.UserController.CreateSession(state.Interaction, ir)
		if err != nil {
			return nil, err
		}

		return &WebAppResult{
			Cookies: result.Cookies,
		}, nil
	default:
		panic("interaction_flow_webapp: unexpected step " + s.CurrentStep().Step)
	}
}
