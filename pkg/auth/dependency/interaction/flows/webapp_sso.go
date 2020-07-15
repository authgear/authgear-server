package flows

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/interaction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/sso"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/core/crypto"
)

type OAuthCallbackData struct {
	State            string
	Code             string
	Scope            string
	Error            string
	ErrorDescription string
}

// OAuthRedirectError wraps err and redirectURI.
// Its purpose is to instruct the error handler to use the provided redirectURI.
type OAuthRedirectError struct {
	redirectURI string
	err         error
}

func (e *OAuthRedirectError) Error() string {
	return e.err.Error()
}

func (e *OAuthRedirectError) RedirectURI() string {
	return e.redirectURI
}

func (e *OAuthRedirectError) Unwrap() error {
	return e.err
}

type OAuthAction string

const (
	OAuthActionLogin   OAuthAction = "login"
	OAuthActionLink    OAuthAction = "link"
	OAuthActionPromote OAuthAction = "promote"
)

type BeginOAuthOptions struct {
	ProviderAlias string
	Action        OAuthAction
	UserID        string
	NonceSource   *http.Cookie
}

func (f *WebAppFlow) BeginOAuth(state *State, opts BeginOAuthOptions) (result *WebAppResult, err error) {
	oauthProvider := f.OAuthProviderFactory.NewOAuthProvider(opts.ProviderAlias)
	if oauthProvider == nil {
		err = ErrOAuthProviderNotFound
		return
	}

	if opts.NonceSource == nil || opts.NonceSource.Value == "" {
		err = errors.New("webapp: failed to generate nonce")
		return
	}

	nonce := crypto.SHA256String(opts.NonceSource.Value)

	param := sso.GetAuthURLParam{
		State: state.InstanceID,
		Nonce: nonce,
	}

	authURI, err := oauthProvider.GetAuthURL(param)
	if err != nil {
		return
	}

	providerConfig := oauthProvider.Config()
	providerID := providerConfig.ProviderID()

	identitySpec := identity.Spec{
		Type: authn.IdentityTypeOAuth,
		Claims: map[string]interface{}{
			identity.IdentityClaimOAuthProviderKeys:                 providerID.Claims(),
			identity.IdentityClaimOAuthAction:                       string(opts.Action),
			identity.IdentityClaimOAuthNonce:                        nonce,
			identity.IdentityClaimOAuthUserID:                       opts.UserID,
			identity.IdentityClaimOAuthGeneratedProviderRedirectURI: authURI,
		},
	}

	clientID := ""
	state.Interaction, err = f.Interactions.NewInteractionOAuth(&interaction.IntentOAuth{
		Identity: identitySpec,
	}, clientID)
	if err != nil {
		return
	}

	result = &WebAppResult{}
	return
}

type HandleOAuthCallbackOptions struct {
	ProviderAlias string
	NonceSource   *http.Cookie
}

func (f *WebAppFlow) HandleOAuthCallback(state *State, data OAuthCallbackData, opts HandleOAuthCallbackOptions) (result *WebAppResult, err error) {
	stepState, err := f.Interactions.GetStepState(state.Interaction)
	if err != nil {
		return
	}
	if stepState.Step != interaction.StepOAuth {
		panic(fmt.Sprintf("webapp: unexpected step: %v", stepState.Step))
	}

	action, _ := stepState.Identity.Claims[identity.IdentityClaimOAuthAction].(string)
	userID, _ := stepState.Identity.Claims[identity.IdentityClaimOAuthUserID].(string)
	hashedNonce, _ := stepState.Identity.Claims[identity.IdentityClaimOAuthNonce].(string)
	redirectURI, _ := state.Extra[ExtraRedirectURI].(string)

	// Wrap the error so that we can go back where we were.
	defer func() {
		if err != nil {
			err = &OAuthRedirectError{
				redirectURI: redirectURI,
				err:         err,
			}
		}
	}()

	oauthProvider := f.OAuthProviderFactory.NewOAuthProvider(opts.ProviderAlias)
	if oauthProvider == nil {
		err = ErrOAuthProviderNotFound
		return
	}

	// Handle provider error
	if data.Error != "" {
		msg := "login failed"
		if desc := data.ErrorDescription; desc != "" {
			msg += ": " + desc
		}
		err = sso.NewSSOFailed(sso.SSOUnauthorized, msg)
		return
	}

	// Verify CSRF cookie
	if opts.NonceSource == nil || opts.NonceSource.Value == "" {
		err = sso.NewSSOFailed(sso.SSOUnauthorized, "invalid nonce")
		return
	}
	hashedCookie := crypto.SHA256String(opts.NonceSource.Value)
	if subtle.ConstantTimeCompare([]byte(hashedNonce), []byte(hashedCookie)) != 1 {
		err = sso.NewSSOFailed(sso.SSOUnauthorized, "invalid nonce")
		return
	}

	oauthAuthInfo, err := oauthProvider.GetAuthInfo(
		sso.OAuthAuthorizationResponse{
			Code:  data.Code,
			State: data.State,
			Scope: data.Scope,
		},
		sso.GetAuthInfoParam{
			Nonce: hashedNonce,
		},
	)
	if err != nil {
		return
	}

	switch OAuthAction(action) {
	case OAuthActionLogin:
		result, err = f.loginWithOAuthProvider(state, oauthAuthInfo)
	case OAuthActionLink:
		result, err = f.linkWithOAuthProvider(state, userID, oauthAuthInfo)
	case OAuthActionPromote:
		result, err = f.promoteWithOAuthProvider(state, userID, oauthAuthInfo)
	default:
		panic(fmt.Errorf("webapp: unexpected sso action: %v", action))
	}

	return
}

func (f *WebAppFlow) loginWithOAuthProvider(state *State, oauthAuthInfo sso.AuthInfo) (*WebAppResult, error) {
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

func (f *WebAppFlow) linkWithOAuthProvider(state *State, userID string, oauthAuthInfo sso.AuthInfo) (result *WebAppResult, err error) {
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
