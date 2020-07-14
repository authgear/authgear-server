package webapp

import (
	"crypto/subtle"
	"fmt"
	"net/http"

	interactionflows "github.com/authgear/authgear-server/pkg/auth/dependency/interaction/flows"
	"github.com/authgear/authgear-server/pkg/auth/dependency/sso"
	"github.com/authgear/authgear-server/pkg/core/crypto"
	"github.com/authgear/authgear-server/pkg/core/errors"
)

type OAuthProviderFactory interface {
	NewOAuthProvider(alias string) sso.OAuthProvider
}

type OAuthInteractions interface {
	LoginWithOAuthProvider(state *interactionflows.State, ssoAuthInfo sso.AuthInfo) (*interactionflows.WebAppResult, error)
	LinkWithOAuthProvider(state *interactionflows.State, userID string, ssoAuthInfo sso.AuthInfo) (*interactionflows.WebAppResult, error)
	PromoteWithOAuthProvider(state *interactionflows.State, userID string, ssoAuthInfo sso.AuthInfo) (*interactionflows.WebAppResult, error)
}

type OAuthService struct {
	OAuthProviderFactory OAuthProviderFactory
	Interactions         OAuthInteractions
}

type SSOCallbackData struct {
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

func (s *OAuthService) handleOAuth(
	r *http.Request,
	providerAlias string,
	action string,
	userID string,
	state *interactionflows.State,
) (result *interactionflows.WebAppResult, err error) {
	var authURI string

	oauthProvider := s.OAuthProviderFactory.NewOAuthProvider(providerAlias)
	if oauthProvider == nil {
		err = ErrOAuthProviderNotFound
		return
	}

	// Use the CSRF token as nonce
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil || cookie.Value == "" {
		panic(errors.Newf("webapp: missing csrf cookies: %w", err))
	}

	nonce := crypto.SHA256String(cookie.Value)

	param := sso.GetAuthURLParam{
		State: state.ID,
		Nonce: nonce,
	}

	authURI, err = oauthProvider.GetAuthURL(param)
	if err != nil {
		return
	}

	state.Extra[interactionflows.ExtraSSOAction] = action
	state.Extra[interactionflows.ExtraSSOUserID] = userID
	state.Extra[interactionflows.ExtraSSONonce] = nonce

	result = &interactionflows.WebAppResult{
		RedirectURI: authURI,
	}
	return
}

func (s *OAuthService) LoginOAuthProvider(r *http.Request, providerAlias string, state *interactionflows.State) (result *interactionflows.WebAppResult, err error) {
	return s.handleOAuth(r, providerAlias, "login", "", state)
}

func (s *OAuthService) LinkOAuthProvider(r *http.Request, providerAlias string, userID string, state *interactionflows.State) (result *interactionflows.WebAppResult, err error) {
	return s.handleOAuth(r, providerAlias, "link", userID, state)
}

func (s *OAuthService) PromoteOAuthProvider(r *http.Request, providerAlias string, userID string, state *interactionflows.State) (result *interactionflows.WebAppResult, err error) {
	return s.handleOAuth(r, providerAlias, "promote", userID, state)
}

func (s *OAuthService) HandleSSOCallback(r *http.Request, providerAlias string, state *interactionflows.State, data SSOCallbackData) (result *interactionflows.WebAppResult, err error) {
	action, _ := state.Extra[interactionflows.ExtraSSOAction].(string)
	userID, _ := state.Extra[interactionflows.ExtraSSOUserID].(string)
	redirectURI, _ := state.Extra[interactionflows.ExtraRedirectURI].(string)

	// Wrap the error so that we can go back where we were.
	defer func() {
		if err != nil {
			err = &OAuthRedirectError{
				redirectURI: redirectURI,
				err:         err,
			}
		}
	}()

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
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil || cookie.Value == "" {
		err = sso.NewSSOFailed(sso.SSOUnauthorized, "invalid nonce")
		return
	}
	hashedCookie := crypto.SHA256String(cookie.Value)
	hashedNonce, ok := state.Extra[interactionflows.ExtraSSONonce].(string)
	if !ok || subtle.ConstantTimeCompare([]byte(hashedNonce), []byte(hashedCookie)) != 1 {
		err = sso.NewSSOFailed(sso.SSOUnauthorized, "invalid nonce")
		return
	}

	oauthProvider := s.OAuthProviderFactory.NewOAuthProvider(providerAlias)
	if oauthProvider == nil {
		err = ErrOAuthProviderNotFound
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

	switch action {
	case "login":
		result, err = s.Interactions.LoginWithOAuthProvider(state, oauthAuthInfo)
	case "link":
		result, err = s.Interactions.LinkWithOAuthProvider(state, userID, oauthAuthInfo)
	case "promote":
		result, err = s.Interactions.PromoteWithOAuthProvider(state, userID, oauthAuthInfo)
	default:
		panic(fmt.Errorf("webapp: unexpected sso action: %v", action))
	}

	return
}
