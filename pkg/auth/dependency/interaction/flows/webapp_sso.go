package flows

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/interaction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/sso"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func (f *WebAppFlow) LoginWithOAuthProvider(state *State, oauthAuthInfo sso.AuthInfo) (*WebAppResult, error) {
	providerID := oauthAuthInfo.ProviderConfig.ProviderID()
	claims := map[string]interface{}{
		identity.IdentityClaimOAuthProviderKeys: providerID.Claims(),
		identity.IdentityClaimOAuthSubjectID:    oauthAuthInfo.ProviderUserInfo.ID,
		identity.IdentityClaimOAuthProfile:      oauthAuthInfo.ProviderRawProfile,
		identity.IdentityClaimOAuthClaims:       oauthAuthInfo.ProviderUserInfo.ClaimsValue(),
	}

	var err error
	state.Interaction, err = f.Interactions.NewInteractionLogin(&interaction.IntentLogin{
		Identity: identity.Spec{
			Type:   authn.IdentityTypeOAuth,
			Claims: claims,
		},
	}, "")

	if err == nil {
		return f.afterPrimaryAuthentication(state)
	}
	if !errors.Is(err, interaction.ErrInvalidCredentials) {
		return nil, err
	}

	// try signup
	state.Interaction, err = f.Interactions.NewInteractionSignup(&interaction.IntentSignup{
		Identity: identity.Spec{
			Type:   authn.IdentityTypeOAuth,
			Claims: claims,
		},
	}, "")
	if err != nil {
		return nil, err
	}
	stepState, err := f.Interactions.GetStepState(state.Interaction)
	if err != nil {
		return nil, err
	}
	if stepState.Step != interaction.StepCommit {
		panic("interaction_flow_webapp: unexpected interaction state")
	}
	result, err := f.Interactions.Commit(state.Interaction)
	if err != nil {
		return nil, err
	}

	// create new interaction after signup
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

	return f.afterPrimaryAuthentication(state)
}

func (f *WebAppFlow) LinkWithOAuthProvider(state *State, userID string, oauthAuthInfo sso.AuthInfo) (result *WebAppResult, err error) {
	providerID := oauthAuthInfo.ProviderConfig.ProviderID()
	claims := map[string]interface{}{
		identity.IdentityClaimOAuthProviderKeys: providerID.Claims(),
		identity.IdentityClaimOAuthSubjectID:    oauthAuthInfo.ProviderUserInfo.ID,
		identity.IdentityClaimOAuthProfile:      oauthAuthInfo.ProviderRawProfile,
		identity.IdentityClaimOAuthClaims:       oauthAuthInfo.ProviderUserInfo.ClaimsValue(),
	}

	clientID := ""
	state.Interaction, err = f.Interactions.NewInteractionAddIdentity(&interaction.IntentAddIdentity{
		Identity: identity.Spec{
			Type:   authn.IdentityTypeOAuth,
			Claims: claims,
		},
	}, clientID, userID)
	if err != nil {
		return
	}

	stepState, err := f.Interactions.GetStepState(state.Interaction)
	if err != nil {
		return
	}

	if stepState.Step != interaction.StepCommit {
		// authenticator is not needed for oauth identity
		// so the current step must be commit
		panic("interaction_flow_webapp: unexpected interaction step")
	}

	_, err = f.Interactions.Commit(state.Interaction)
	if err != nil {
		return nil, err
	}

	result = &WebAppResult{}
	return
}

func (f *WebAppFlow) UnlinkOAuthProvider(state *State, providerAlias string, userID string) (result *WebAppResult, err error) {
	providerConfig, ok := f.SSOOAuthConfig.GetProviderConfig(providerAlias)
	if !ok {
		err = ErrOAuthProviderNotFound
		return
	}

	providerID := providerConfig.ProviderID()
	clientID := ""
	state.Interaction, err = f.Interactions.NewInteractionRemoveIdentity(&interaction.IntentRemoveIdentity{
		Identity: identity.Spec{
			Type: authn.IdentityTypeOAuth,
			Claims: map[string]interface{}{
				identity.IdentityClaimOAuthProviderKeys: providerID.Claims(),
			},
		},
	}, clientID, userID)
	if err != nil {
		return
	}

	stepState, err := f.Interactions.GetStepState(state.Interaction)
	if err != nil {
		return
	}

	if stepState.Step != interaction.StepCommit {
		panic("interaction_flow_webapp: unexpected step " + stepState.Step)
	}

	_, err = f.Interactions.Commit(state.Interaction)
	if err != nil {
		return
	}

	result = &WebAppResult{}
	return
}
