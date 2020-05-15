package flows

import (
	"errors"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func (f *WebAppFlow) LoginWithOAuthProvider(oauthAuthInfo sso.AuthInfo) (*WebAppResult, error) {
	providerID := oauth.NewProviderID(oauthAuthInfo.ProviderConfig)
	claims := map[string]interface{}{
		identity.IdentityClaimOAuthProvider:  providerID.ClaimsValue(),
		identity.IdentityClaimOAuthSubjectID: oauthAuthInfo.ProviderUserInfo.ID,
		identity.IdentityClaimOAuthProfile:   oauthAuthInfo.ProviderRawProfile,
		identity.IdentityClaimOAuthClaims:    oauthAuthInfo.ProviderUserInfo.ClaimsValue(),
	}
	i, err := f.Interactions.NewInteractionLogin(&interaction.IntentLogin{
		Identity: identity.Spec{
			Type:   authn.IdentityTypeOAuth,
			Claims: claims,
		},
	}, "")
	if err == nil {
		return f.afterPrimaryAuthentication(i)
	}
	if !errors.Is(err, interaction.ErrInvalidCredentials) {
		return nil, err
	}

	// try signup
	i, err = f.Interactions.NewInteractionSignup(&interaction.IntentSignup{
		Identity: identity.Spec{
			Type:   authn.IdentityTypeOAuth,
			Claims: claims,
		},
		OnUserDuplicate: model.OnUserDuplicateAbort,
	}, "")
	if err != nil {
		return nil, err
	}
	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return nil, err
	}
	if s.CurrentStep().Step != interaction.StepCommit {
		panic("interaction_flow_webapp: unexpected interaction state")
	}
	attrs, err := f.Interactions.Commit(i)
	if err != nil {
		return nil, err
	}

	// create new interaction after signup
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
	return f.afterPrimaryAuthentication(i)
}

func (f *WebAppFlow) LinkWithOAuthProvider(userID string, oauthAuthInfo sso.AuthInfo) (result *WebAppResult, err error) {
	providerID := oauth.NewProviderID(oauthAuthInfo.ProviderConfig)
	claims := map[string]interface{}{
		identity.IdentityClaimOAuthProvider:  providerID.ClaimsValue(),
		identity.IdentityClaimOAuthSubjectID: oauthAuthInfo.ProviderUserInfo.ID,
		identity.IdentityClaimOAuthProfile:   oauthAuthInfo.ProviderRawProfile,
		identity.IdentityClaimOAuthClaims:    oauthAuthInfo.ProviderUserInfo.ClaimsValue(),
	}

	clientID := ""
	i, err := f.Interactions.NewInteractionAddIdentity(&interaction.IntentAddIdentity{
		Identity: identity.Spec{
			Type:   authn.IdentityTypeOAuth,
			Claims: claims,
		},
	}, clientID, userID)
	if err != nil {
		return
	}

	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return
	}

	if s.CurrentStep().Step != interaction.StepCommit {
		// authenticator is not needed for oauth identity
		// so the current step must be commit
		panic("interaction_flow_webapp: unexpected interaction step")
	}

	_, err = f.Interactions.Commit(i)
	if err != nil {
		return nil, err
	}

	result = &WebAppResult{
		Step: WebAppStepCompleted,
	}

	return
}

func (f *WebAppFlow) UnlinkWithOAuthProvider(userID string, providerConfig config.OAuthProviderConfiguration) (result *WebAppResult, err error) {
	providerID := oauth.NewProviderID(providerConfig)
	clientID := ""
	i, err := f.Interactions.NewInteractionRemoveIdentity(&interaction.IntentRemoveIdentity{
		Identity: identity.Spec{
			Type: authn.IdentityTypeOAuth,
			Claims: map[string]interface{}{
				identity.IdentityClaimOAuthProvider: providerID.ClaimsValue(),
			},
		},
	}, clientID, userID)
	if err != nil {
		return
	}

	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return
	}

	if s.CurrentStep().Step != interaction.StepCommit {
		panic("interaction_flow_webapp: unexpected step " + s.CurrentStep().Step)
	}

	_, err = f.Interactions.Commit(i)
	if err != nil {
		return
	}

	result = &WebAppResult{
		Step: WebAppStepCompleted,
	}

	return
}

func (f *WebAppFlow) PromoteWithOAuthProvider(userID string, oauthAuthInfo sso.AuthInfo) (*WebAppResult, error) {
	providerID := oauth.NewProviderID(oauthAuthInfo.ProviderConfig)
	claims := map[string]interface{}{
		identity.IdentityClaimOAuthProvider:  providerID.ClaimsValue(),
		identity.IdentityClaimOAuthSubjectID: oauthAuthInfo.ProviderUserInfo.ID,
		identity.IdentityClaimOAuthProfile:   oauthAuthInfo.ProviderRawProfile,
		identity.IdentityClaimOAuthClaims:    oauthAuthInfo.ProviderUserInfo.ClaimsValue(),
	}

	clientID := ""
	i, err := f.Interactions.NewInteractionAddIdentity(&interaction.IntentAddIdentity{
		Identity: identity.Spec{
			Type:   authn.IdentityTypeOAuth,
			Claims: claims,
		},
	}, clientID, userID)
	if err != nil {
		return nil, err
	}

	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return nil, err
	} else if s.CurrentStep().Step != interaction.StepCommit {
		// authenticator is not needed for oauth identity
		// so the current step must be commit
		panic("interaction_flow_webapp: unexpected interaction step")
	}

	attrs, err := f.Interactions.Commit(i)
	if err != nil {
		return nil, err
	}

	return f.afterAnonymousUserPromotion(attrs)
}
