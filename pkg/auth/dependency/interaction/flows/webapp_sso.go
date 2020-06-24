package flows

import (
	"errors"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

func (f *WebAppFlow) LoginWithOAuthProvider(oauthAuthInfo sso.AuthInfo) (*WebAppResult, error) {
	providerID := oauthAuthInfo.ProviderConfig.ProviderID()
	claims := map[string]interface{}{
		identity.IdentityClaimOAuthProviderKeys: providerID.Claims(),
		identity.IdentityClaimOAuthSubjectID:    oauthAuthInfo.ProviderUserInfo.ID,
		identity.IdentityClaimOAuthProfile:      oauthAuthInfo.ProviderRawProfile,
		identity.IdentityClaimOAuthClaims:       oauthAuthInfo.ProviderUserInfo.ClaimsValue(),
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
	result, err := f.Interactions.Commit(i)
	if err != nil {
		return nil, err
	}

	// create new interaction after signup
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
	return f.afterPrimaryAuthentication(i)
}

func (f *WebAppFlow) LinkWithOAuthProvider(userID string, oauthAuthInfo sso.AuthInfo) (result *WebAppResult, err error) {
	providerID := oauthAuthInfo.ProviderConfig.ProviderID()
	claims := map[string]interface{}{
		identity.IdentityClaimOAuthProviderKeys: providerID.Claims(),
		identity.IdentityClaimOAuthSubjectID:    oauthAuthInfo.ProviderUserInfo.ID,
		identity.IdentityClaimOAuthProfile:      oauthAuthInfo.ProviderRawProfile,
		identity.IdentityClaimOAuthClaims:       oauthAuthInfo.ProviderUserInfo.ClaimsValue(),
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

func (f *WebAppFlow) UnlinkWithOAuthProvider(userID string, providerConfig *config.OAuthSSOProviderConfig) (result *WebAppResult, err error) {
	providerID := providerConfig.ProviderID()
	clientID := ""
	i, err := f.Interactions.NewInteractionRemoveIdentity(&interaction.IntentRemoveIdentity{
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
