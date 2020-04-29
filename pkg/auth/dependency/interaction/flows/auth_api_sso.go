package flows

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

const (
	// AuthAPIStateOAuthCodeChallenge is a claim with string value for sso code challenge of current interaction in auth api
	AuthAPIStateOAuthCodeChallenge string = "https://auth.skygear.io/claims/auth_api/sso/code_challenge"
)

func (f *AuthAPIFlow) LoginWithOAuthProvider(
	clientID string, oauthAuthInfo sso.AuthInfo, codeChallenge string, onUserDuplicate model.OnUserDuplicate,
) (string, error) {
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
	}, clientID)
	if err == nil {
		i.State[AuthAPIStateOAuthCodeChallenge] = codeChallenge
		return f.Interactions.SaveInteraction(i)
	}
	if !errors.Is(err, interaction.ErrInvalidCredentials) {
		return "", err
	}

	// try signup
	i, err = f.Interactions.NewInteractionSignup(&interaction.IntentSignup{
		Identity: interaction.IdentitySpec{
			Type:   authn.IdentityTypeOAuth,
			Claims: claims,
		},
		OnUserDuplicate: onUserDuplicate,
	}, clientID)
	if err != nil {
		return "", err
	}
	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return "", err
	}
	if s.CurrentStep().Step != interaction.StepCommit {
		panic("interaction_flow_auth_api: unexpected interaction state")
	}
	_, err = f.Interactions.Commit(i)
	if err != nil {
		return "", err
	}

	// create new interaction after signup
	i, err = f.Interactions.NewInteractionLogin(&interaction.IntentLogin{
		Identity: i.Identity.ToSpec(),
		AuthenticatedAs: &interaction.IntentLoginAuthenticatedAs{
			UserID: i.UserID,
		},
		OriginalIntentType: i.Intent.Type(),
	}, clientID)
	if err != nil {
		return "", err
	}
	i.State[AuthAPIStateOAuthCodeChallenge] = codeChallenge
	return f.Interactions.SaveInteraction(i)
}

func (f *AuthAPIFlow) ExchangeCode(interactionToken string, verifier string) (*AuthResult, error) {
	i, err := f.Interactions.GetInteraction(interactionToken)
	if err != nil {
		return nil, err
	}

	challenge := i.State[AuthAPIStateOAuthCodeChallenge]
	// challenge can be empty for api login with access token flow
	if challenge != "" {
		if err := verifyPKCE(challenge, verifier); err != nil {
			return nil, err
		}
	}

	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return nil, err
	}

	// code verifier

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

func verifyPKCE(challenge string, verifier string) error {
	sha256Arr := sha256.Sum256([]byte(verifier))
	expectedChallenge := base64.RawURLEncoding.EncodeToString(sha256Arr[:])
	if subtle.ConstantTimeCompare([]byte(challenge), []byte(expectedChallenge)) != 1 {
		return sso.NewSSOFailed(sso.InvalidCodeVerifier, "invalid code verifier")
	}
	return nil
}
