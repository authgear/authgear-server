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
	LoginWithOAuthProvider(sso.AuthInfo) (*interactionflows.WebAppResult, error)
	LinkWithOAuthProvider(userID string, ssoAuthInfo sso.AuthInfo) (*interactionflows.WebAppResult, error)
	PromoteWithOAuthProvider(userID string, ssoAuthInfo sso.AuthInfo) (*interactionflows.WebAppResult, error)
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

func (s *OAuthService) LoginOAuthProvider(w http.ResponseWriter, r *http.Request, providerAlias string, state *State) (result *interactionflows.WebAppResult, err error) {
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

	state.Extra[ExtraSSOAction] = "login"
	state.Extra[ExtraSSONonce] = nonce
	state.Extra[ExtraSSORedirectURI] = r.URL.String()

	result = &interactionflows.WebAppResult{
		RedirectURI: authURI,
	}
	return
}

func (s *OAuthService) HandleSSOCallback(r *http.Request, providerAlias string, state *State, data SSOCallbackData) (result *interactionflows.WebAppResult, err error) {
	action, _ := state.Extra[ExtraSSOAction].(string)
	userID, _ := state.Extra[ExtraUserID].(string)
	redirectURI, _ := state.Extra[ExtraSSORedirectURI].(string)

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
	hashedNonce, ok := state.Extra[ExtraSSONonce].(string)
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
		result, err = s.Interactions.LoginWithOAuthProvider(oauthAuthInfo)
	case "link":
		result, err = s.Interactions.LinkWithOAuthProvider(userID, oauthAuthInfo)
	case "promote":
		result, err = s.Interactions.PromoteWithOAuthProvider(userID, oauthAuthInfo)
	default:
		panic(fmt.Errorf("webapp: unexpected sso action: %v", action))
	}

	return
}
