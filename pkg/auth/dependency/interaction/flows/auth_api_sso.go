package flows

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

const (
	// AuthAPIExtraStateOAuthCodeChallenge is a extra state with string value for sso code challenge of current interaction in auth api
	AuthAPIExtraStateOAuthCodeChallenge string = "https://auth.skygear.io/claims/auth_api/sso/code_challenge"
)

func (f *AuthAPIFlow) LoginWithOAuthProvider(
	clientID string, oauthAuthInfo sso.AuthInfo, codeChallenge string, onUserDuplicate model.OnUserDuplicate,
) (string, error) {
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
	}, clientID)
	if err == nil {
		i.Extra[AuthAPIExtraStateOAuthCodeChallenge] = codeChallenge
		return f.Interactions.SaveInteraction(i)
	}
	if !errors.Is(err, interaction.ErrInvalidCredentials) {
		return "", err
	}

	// try signup
	i, err = f.Interactions.NewInteractionSignup(&interaction.IntentSignup{
		Identity: identity.Spec{
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
	attrs, err := f.Interactions.Commit(i)
	if err != nil {
		return "", err
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
		return "", err
	}
	i.Extra[AuthAPIExtraStateOAuthCodeChallenge] = codeChallenge
	return f.Interactions.SaveInteraction(i)
}

func (f *AuthAPIFlow) LinkWithOAuthProvider(
	clientID string, userID string, oauthAuthInfo sso.AuthInfo, codeChallenge string,
) (string, error) {
	providerID := oauth.NewProviderID(oauthAuthInfo.ProviderConfig)
	claims := map[string]interface{}{
		identity.IdentityClaimOAuthProvider:  providerID.ClaimsValue(),
		identity.IdentityClaimOAuthSubjectID: oauthAuthInfo.ProviderUserInfo.ID,
		identity.IdentityClaimOAuthProfile:   oauthAuthInfo.ProviderRawProfile,
		identity.IdentityClaimOAuthClaims:    oauthAuthInfo.ProviderUserInfo.ClaimsValue(),
	}
	i, err := f.Interactions.NewInteractionAddIdentity(&interaction.IntentAddIdentity{
		Identity: identity.Spec{
			Type:   authn.IdentityTypeOAuth,
			Claims: claims,
		},
	}, clientID, userID)
	if err != nil {
		if errors.Is(err, interaction.ErrDuplicatedIdentity) {
			return "", sso.NewSSOFailed(sso.AlreadyLinked, "user is already linked to this provider")
		}
		return "", err
	}

	s, err := f.Interactions.GetInteractionState(i)
	if err != nil {
		return "", err
	} else if s.CurrentStep().Step != interaction.StepCommit {
		// authenticator is not needed for oauth identity
		// so the current step must be commit
		panic("interaction_flow_auth_api: unexpected interaction step")
	}

	i.Extra[AuthAPIExtraStateOAuthCodeChallenge] = codeChallenge
	return f.Interactions.SaveInteraction(i)
}

func (f *AuthAPIFlow) ExchangeCode(interactionToken string, verifier string) (*AuthResult, error) {
	i, err := f.Interactions.GetInteraction(interactionToken)
	if err != nil {
		return nil, err
	}

	challenge := i.Extra[AuthAPIExtraStateOAuthCodeChallenge]
	// challenge can be empty for api login with access token flow
	if challenge != "" {
		if err := verifyPKCE(challenge, verifier); err != nil {
			return nil, err
		}
	}

	if _, ok := i.Intent.(*interaction.IntentAddIdentity); ok {
		// link
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

	s, err := f.Interactions.GetInteractionState(i)
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

func (f *AuthAPIFlow) UnlinkkWithOAuthProvider(
	clientID string, userID string, oauthProviderInfo config.OAuthProviderConfiguration,
) error {
	providerID := oauth.NewProviderID(oauthProviderInfo)
	i, err := f.Interactions.NewInteractionRemoveIdentity(&interaction.IntentRemoveIdentity{
		Identity: identity.Spec{
			Type: authn.IdentityTypeOAuth,
			Claims: map[string]interface{}{
				identity.IdentityClaimOAuthProvider: providerID.ClaimsValue(),
			},
		},
	}, clientID, userID)
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

func verifyPKCE(challenge string, verifier string) error {
	sha256Arr := sha256.Sum256([]byte(verifier))
	expectedChallenge := base64.RawURLEncoding.EncodeToString(sha256Arr[:])
	if subtle.ConstantTimeCompare([]byte(challenge), []byte(expectedChallenge)) != 1 {
		return sso.NewSSOFailed(sso.InvalidCodeVerifier, "invalid code verifier")
	}
	return nil
}
