package flows

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
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
		Identity: identity.Spec{
			Type: authn.IdentityTypeLoginID,
			Claims: map[string]interface{}{
				identity.IdentityClaimLoginIDKey:   loginIDKey,
				identity.IdentityClaimLoginIDValue: loginID,
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
			Authenticator: authenticator.Spec{Type: authn.AuthenticatorTypePassword},
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
		Identity: identity.Spec{
			Type: authn.IdentityTypeLoginID,
			Claims: map[string]interface{}{
				identity.IdentityClaimLoginIDKey:   loginIDKey,
				identity.IdentityClaimLoginIDValue: loginID,
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
			Authenticator: authenticator.Spec{Type: authn.AuthenticatorTypePassword},
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
		Identity: identity.Spec{
			Type: authn.IdentityTypeLoginID,
			Claims: map[string]interface{}{
				identity.IdentityClaimLoginIDKey:   loginIDKey,
				identity.IdentityClaimLoginIDValue: loginID,
			},
		},
	}, session.GetClientID(), session.AuthnAttrs().UserID)
	if err != nil {
		return err
	}

	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return err
	}

	// in auth api, only password authenticator is supported for login id
	// if the step is StepSetupPrimaryAuthenticator, it must be password authenticator
	// if user has password authenticator already, the step should be commit
	ss := s.CurrentStep()
	if ss.Step == interaction.StepSetupPrimaryAuthenticator &&
		len(ss.AvailableAuthenticators) > 0 &&
		ss.AvailableAuthenticators[0].Type == authn.AuthenticatorTypePassword {
		passwordAuthenticator := ss.AvailableAuthenticators[0]

		// Set password authenticator to no password
		// Before resetting the password, user cannot use this authenticator to authenticate
		err = f.Interactions.PerformAction(i, interaction.StepSetupPrimaryAuthenticator, &interaction.ActionSetupAuthenticator{
			Authenticator: passwordAuthenticator,
			Secret:        "",
		})
		if err != nil {
			return err
		}

		if i.Error != nil {
			return i.Error
		}

		s, err = f.Interactions.GetInteractionState(i)
		if err != nil {
			return err
		}
	}

	if s.CurrentStep().Step != interaction.StepCommit {
		return ErrUnsupportedConfiguration
	}

	_, err = f.Interactions.Commit(i)
	if err != nil {
		return err
	}

	return nil
}

func (f *AuthAPIFlow) RemoveLoginID(
	loginIDKey string, loginID string, session auth.AuthSession,
) error {
	i, err := f.Interactions.NewInteractionRemoveIdentity(&interaction.IntentRemoveIdentity{
		Identity: identity.Spec{
			Type: authn.IdentityTypeLoginID,
			Claims: map[string]interface{}{
				identity.IdentityClaimLoginIDKey:   loginIDKey,
				identity.IdentityClaimLoginIDValue: loginID,
			},
		},
	}, session.GetClientID(), session.AuthnAttrs().UserID)
	if err != nil {
		return err
	}

	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return err
	}

	if s.CurrentStep().Step != interaction.StepCommit {
		panic("interaction_flow_webapp: unexpected step " + s.CurrentStep().Step)
	}

	_, err = f.Interactions.Commit(i)
	if err != nil {
		return err
	}

	return nil
}

func (f *AuthAPIFlow) UpdateLoginID(
	oldLoginID loginid.LoginID, newLoginID loginid.LoginID, session auth.AuthSession,
) (*AuthResult, error) {
	i, err := f.Interactions.NewInteractionUpdateIdentity(&interaction.IntentUpdateIdentity{
		OldIdentity: interaction.IdentitySpec{
			Type: authn.IdentityTypeLoginID,
			Claims: map[string]interface{}{
				interaction.IdentityClaimLoginIDKey:   oldLoginID.Key,
				interaction.IdentityClaimLoginIDValue: oldLoginID.Value,
			},
		},
		NewIdentity: interaction.IdentitySpec{
			Type: authn.IdentityTypeLoginID,
			Claims: map[string]interface{}{
				interaction.IdentityClaimLoginIDKey:   newLoginID.Key,
				interaction.IdentityClaimLoginIDValue: newLoginID.Value,
			},
		},
	}, session.GetClientID(), session.AuthnAttrs().UserID)
	if err != nil {
		return nil, err
	}

	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return nil, err
	}

	// in auth api, only password authenticator is supported for login id
	// so user should have password authenticator for the original identity
	// already and should not need to setup new primary authenticator
	// current step must be commit
	if s.CurrentStep().Step != interaction.StepCommit {
		return nil, ErrUnsupportedConfiguration
	}

	attrs, err := f.Interactions.Commit(i)
	if err != nil {
		return nil, err
	}
	result, err := f.UserController.MakeAuthResult(attrs)
	if err != nil {
		return nil, err
	}
	return result, nil
}
