package flows

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/interaction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/sso"
	"github.com/authgear/authgear-server/pkg/auth/event"
	"github.com/authgear/authgear-server/pkg/auth/model"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/core/errors"
)

func (f *WebAppFlow) PromoteWithLoginID(state *State, loginIDKey, loginID string, userID string) (*WebAppResult, error) {
	var err error

	iden := identity.Spec{
		Type: authn.IdentityTypeLoginID,
		Claims: map[string]interface{}{
			identity.IdentityClaimLoginIDKey:   loginIDKey,
			identity.IdentityClaimLoginIDValue: loginID,
		},
	}

	if f.Config.OnConflict.Promotion == config.PromotionConflictBehaviorLogin {
		_, _, err = f.Identities.GetByClaims(authn.IdentityTypeLoginID, iden.Claims)
		if errors.Is(err, identity.ErrIdentityNotFound) {
			state.Interaction, err = f.Interactions.NewInteractionAddIdentity(&interaction.IntentAddIdentity{
				Identity: iden,
			}, "", userID)
		} else if err != nil {
			return nil, err
		} else {
			state.Interaction, err = f.Interactions.NewInteractionLogin(&interaction.IntentLogin{
				Identity: iden,
			}, "")
		}
	} else {
		state.Interaction, err = f.Interactions.NewInteractionAddIdentity(&interaction.IntentAddIdentity{
			Identity: iden,
		}, "", userID)
	}
	if err != nil {
		return nil, err
	}

	if state.Interaction.Intent.Type() == interaction.IntentTypeLogin {
		err = f.handleLogin(state)
	} else {
		err = f.handleSignup(state)
	}
	if err != nil {
		return nil, err
	}

	state.Extra[ExtraAnonymousUserID] = userID
	state.Extra[ExtraGivenLoginID] = loginID

	return &WebAppResult{}, nil
}

func (f *WebAppFlow) PromoteWithOAuthProvider(state *State, userID string, oauthAuthInfo sso.AuthInfo) (*WebAppResult, error) {
	providerID := oauthAuthInfo.ProviderConfig.ProviderID()
	iden := identity.Spec{
		Type: authn.IdentityTypeOAuth,
		Claims: map[string]interface{}{
			identity.IdentityClaimOAuthProviderKeys: providerID.Claims(),
			identity.IdentityClaimOAuthSubjectID:    oauthAuthInfo.ProviderUserInfo.ID,
			identity.IdentityClaimOAuthProfile:      oauthAuthInfo.ProviderRawProfile,
			identity.IdentityClaimOAuthClaims:       oauthAuthInfo.ProviderUserInfo.ClaimsValue(),
		},
	}
	var err error

	if f.Config.OnConflict.Promotion == config.PromotionConflictBehaviorLogin {
		_, _, err = f.Identities.GetByClaims(authn.IdentityTypeOAuth, iden.Claims)
		if errors.Is(err, identity.ErrIdentityNotFound) {
			state.Interaction, err = f.Interactions.NewInteractionAddIdentity(&interaction.IntentAddIdentity{
				Identity: iden,
			}, "", userID)
		} else if err != nil {
			return nil, err
		} else {
			state.Interaction, err = f.Interactions.NewInteractionLogin(&interaction.IntentLogin{
				Identity: iden,
			}, "")
		}
	} else {
		state.Interaction, err = f.Interactions.NewInteractionAddIdentity(&interaction.IntentAddIdentity{
			Identity: iden,
		}, "", userID)
	}
	if err != nil {
		return nil, err
	}

	stepState, err := f.Interactions.GetStepState(state.Interaction)
	if err != nil {
		return nil, err
	} else if stepState.Step != interaction.StepCommit {
		// authenticator is not needed for oauth identity
		// so the current step must be commit
		panic("interaction_flow_webapp: unexpected interaction step")
	}

	state.Extra[ExtraAnonymousUserID] = userID

	result, err := f.Interactions.Commit(state.Interaction)
	if err != nil {
		return nil, err
	}

	return f.afterAnonymousUserPromotion(state, result)
}

func (f *WebAppFlow) afterAnonymousUserPromotion(state *State, ir *interaction.Result) (*WebAppResult, error) {
	var err error
	anonUserID, _ := state.Extra[ExtraAnonymousUserID].(string)

	anonUser, err := f.Users.Get(anonUserID)
	if err != nil {
		return nil, err
	}

	// Remove anonymous identity if the same user is reused
	if anonUserID == ir.Attrs.UserID {
		state.Interaction, err = f.Interactions.NewInteractionRemoveIdentity(&interaction.IntentRemoveIdentity{
			Identity: identity.Spec{
				Type:   authn.IdentityTypeAnonymous,
				Claims: map[string]interface{}{},
			},
		}, "", anonUserID)
		if err != nil {
			return nil, err
		}

		stepState, err := f.Interactions.GetStepState(state.Interaction)
		if err != nil {
			return nil, err
		}

		if stepState.Step != interaction.StepCommit {
			panic("interaction_flow_webapp: unexpected step " + stepState.Step)
		}

		_, err = f.Interactions.Commit(state.Interaction)
		if err != nil {
			return nil, err
		}
	}

	user, err := f.Users.Get(ir.Attrs.UserID)
	if err != nil {
		return nil, err
	}

	err = f.Hooks.DispatchEvent(
		event.UserPromoteEvent{
			AnonymousUser: *anonUser,
			User:          *user,
			Identities: []model.Identity{
				ir.Identity.ToModel(),
			},
		},
		user,
	)
	if err != nil {
		return nil, err
	}

	result, err := f.UserController.CreateSession(state.Interaction, ir)
	if err != nil {
		return nil, err
	}

	// NOTE: existing anonymous sessions are not deleted, in case of commit
	// failure may cause lost users.

	return &WebAppResult{
		Cookies: result.Cookies,
	}, nil
}
