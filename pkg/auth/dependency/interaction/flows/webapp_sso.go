package flows

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

func (f *WebAppFlow) LoginWithOAuthProvider(oauthAuthInfo sso.AuthInfo) (*WebAppResult, error) {
	providerID := oauth.NewProviderID(oauthAuthInfo.ProviderConfig)
	claims := map[string]interface{}{
		interaction.IdentityClaimOAuthProvider:  providerID.ClaimsValue(),
		interaction.IdentityClaimOAuthSubjectID: oauthAuthInfo.ProviderUserInfo.ID,
		interaction.IdentityClaimOAuthProfile:   oauthAuthInfo.ProviderRawProfile,
		interaction.IdentityClaimOAuthClaims:    oauthAuthInfo.ProviderUserInfo.ClaimsValue(),
	}
	i, err := f.Interactions.NewInteractionLogin(&interaction.IntentLogin{
		Identity: interaction.IdentitySpec{
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

	switch s.CurrentStep().Step {
	case interaction.StepAuthenticateSecondary, interaction.StepSetupSecondaryAuthenticator:
		panic("interaction_flow_webapp: TODO: handle MFA")

	case interaction.StepCommit:
		// TODO(interaction): update user last login / sso profile in login commit
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
