package flows

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

type AuthAPIFlow struct {
	Interactions   InteractionProvider
	UserController *UserController
}

func (f *AuthAPIFlow) LoginWithLoginIDPassword(
	clientID string, loginIDKey string, loginID string, password string,
) (*AuthResult, error) {
	i, err := f.Interactions.NewInteractionLogin(&interaction.IntentLogin{
		Identity: interaction.IdentitySpec{
			Type: authn.IdentityTypeLoginID,
			Claims: map[string]interface{}{
				interaction.IdentityClaimLoginIDKey:   loginIDKey,
				interaction.IdentityClaimLoginIDValue: loginID,
			},
		},
	}, clientID)
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
			Authenticator: interaction.AuthenticatorSpec{Type: authn.AuthenticatorTypePassword},
			Secret:        password,
		},
	)
	if err != nil {
		return nil, err
	}

	if i.Error != nil {
		return nil, i.Error
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

// TODO(interaction): support multiple login IDs
func (f *AuthAPIFlow) SignupWithLoginIDPassword(
	clientID string,
	loginIDKey string,
	loginID string,
	password string,
	metadata map[string]interface{},
	onUserDuplicate model.OnUserDuplicate,
) (*AuthResult, error) {
	i, err := f.Interactions.NewInteractionSignup(&interaction.IntentSignup{
		Identity: interaction.IdentitySpec{
			Type: authn.IdentityTypeLoginID,
			Claims: map[string]interface{}{
				interaction.IdentityClaimLoginIDKey:   loginIDKey,
				interaction.IdentityClaimLoginIDValue: loginID,
			},
		},
		OnUserDuplicate: onUserDuplicate,
		UserMetadata:    metadata,
	}, clientID)
	if err != nil {
		return nil, err
	}

	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return nil, err
	} else if s.CurrentStep().Step != interaction.StepSetupPrimaryAuthenticator {
		return nil, ErrUnsupportedConfiguration
	}

	err = f.Interactions.PerformAction(
		i,
		interaction.StepSetupPrimaryAuthenticator,
		&interaction.ActionSetupAuthenticator{
			Authenticator: interaction.AuthenticatorSpec{Type: authn.AuthenticatorTypePassword},
			Secret:        password,
		},
	)
	if err != nil {
		return nil, err
	}

	if i.Error != nil {
		return nil, i.Error
	}

	attrs, err := f.Interactions.Commit(i)
	if err != nil {
		return nil, err
	}

	// Login with new user:

	i, err = f.Interactions.NewInteractionLoginAs(
		&interaction.IntentLogin{
			Identity: interaction.IdentitySpec{
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

func (f *AuthAPIFlow) AddLoginID(
	loginIDKey string, loginID string, session auth.AuthSession,
) error {
	i, err := f.Interactions.NewInteractionAddIdentity(&interaction.IntentAddIdentity{
		Identity: interaction.IdentitySpec{
			Type: authn.IdentityTypeLoginID,
			Claims: map[string]interface{}{
				interaction.IdentityClaimLoginIDKey:   loginIDKey,
				interaction.IdentityClaimLoginIDValue: loginID,
			},
		},
	}, session.GetClientID(), session.AuthnAttrs().UserID)
	if err != nil {
		return err
	}

	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return err
	} else if s.CurrentStep().Step != interaction.StepCommit {
		// in auth api, only password authenticator is supported for login id
		// password authenticator should be setup during sign up
		// so the current step must be commit
		return ErrUnsupportedConfiguration
	}

	_, err = f.Interactions.Commit(i)
	if err != nil {
		return err
	}

	return nil
}
