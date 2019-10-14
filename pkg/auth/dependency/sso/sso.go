package sso

import (
	"fmt"
	"strings"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/rand"
)

const (
	nonceAlphabet string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// LoginState stores login specific state.
type LoginState struct {
	MergeRealm      string                `json:"merge_realm,omitempty"`
	OnUserDuplicate model.OnUserDuplicate `json:"on_user_duplicate,omitempty"`
}

// LinkState stores link specific state.
type LinkState struct {
	UserID string `json:"user_id,omitempty"`
}

// OAuthAuthorizationCodeFlowState stores OAuth Authorization Code flow state.
type OAuthAuthorizationCodeFlowState struct {
	UXMode      UXMode `json:"ux_mode,omitempty"`
	CallbackURL string `json:"callback_url,omitempty"`
	Action      string `json:"action,omitempty"`
}

// State is an opaque value used by the client to maintain
// state between the request and callback.
// See https://tools.ietf.org/html/rfc6749#section-4.1.1
type State struct {
	LoginState
	LinkState
	OAuthAuthorizationCodeFlowState
	Nonce       string `json:"nonce,omitempty"`
	APIClientID string `json:"api_client_id"`
}

// UXMode indicates how the URL is used
type UXMode string

// UXMode constants
const (
	UXModeWebRedirect UXMode = "web_redirect"
	UXModeWebPopup    UXMode = "web_popup"
	UXModeMobileApp   UXMode = "mobile_app"
)

// IsValidUXMode validates UXMode
func IsValidUXMode(mode UXMode) bool {
	allModes := []UXMode{UXModeWebRedirect, UXModeWebPopup, UXModeMobileApp}
	for _, v := range allModes {
		if mode == v {
			return true
		}
	}
	return false
}

// GetURLParams is the argument of getAuthURL
type GetURLParams struct {
	State State
}

// AuthInfo contains auth info from HandleAuthzResp
type AuthInfo struct {
	ProviderConfig          config.OAuthProviderConfiguration
	ProviderRawProfile      map[string]interface{}
	ProviderAccessTokenResp interface{}
	ProviderUserInfo        ProviderUserInfo
}

type ProviderUserInfo struct {
	ID    string
	Email string
}

type OAuthAuthorizationResponse struct {
	Code  string
	State string
	Scope string
	// Nonce is required when the provider supports OpenID connect or OAuth Authorization Code Flow.
	// The implementation is based on the suggestion in the spec.
	// See https://openid.net/specs/openid-connect-core-1_0.html#NonceNotes
	//
	// The nonce is a cryptographically random string.
	// The nonce is stored in the session cookie when auth URL is called.
	// The nonce is hashed with SHA256.
	// The hashed nonce is given to the OIDC provider
	// The hashed nonce is stored in the state.
	// The callback endpoint expect the user agent to include the nonce in the session cookie.
	// The nonce in session cookie will be validated against the hashed nonce in the ID token.
	// The nonce in session cookie will be validated against the hashed nonce in the state.
	Nonce string
}

// OAuthProvider is OAuth 2.0 based provider.
type OAuthProvider interface {
	GetAuthURL(params GetURLParams) (url string, err error)
	DecodeState(encodedState string) (*State, error)
	GetAuthInfo(r OAuthAuthorizationResponse) (AuthInfo, error)
}

// NonOpenIDConnectProvider are OAuth 2.0 provider that does not
// implement OpenID Connect or we do not implement yet.
// They are Google, Facebook, Instagram and LinkedIn.
type NonOpenIDConnectProvider interface {
	NonOpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse) (authInfo AuthInfo, err error)
}

// ExternalAccessTokenFlowProvider is provider that the developer
// can somehow acquire an access token and that access token
// can be used to fetch user info.
// They are Google, Facebook, Instagram and LinkedIn.
type ExternalAccessTokenFlowProvider interface {
	ExternalAccessTokenGetAuthInfo(AccessTokenResp) (AuthInfo, error)
}

// OpenIDConnectProvider are OpenID Connect provider.
// They are Azure AD v2.
type OpenIDConnectProvider interface {
	OpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse) (authInfo AuthInfo, err error)
}

type ProviderFactory struct {
	urlPrefixProvider urlprefix.Provider
	tenantConfig      config.TenantConfiguration
}

func NewProviderFactory(tenantConfig config.TenantConfiguration, urlPrefixProvider urlprefix.Provider) *ProviderFactory {
	return &ProviderFactory{
		tenantConfig: tenantConfig,
	}
}

func (p *ProviderFactory) NewProvider(id string) OAuthProvider {
	providerConfig, ok := p.tenantConfig.GetOAuthProviderByID(id)
	if !ok {
		return nil
	}
	switch providerConfig.Type {
	case config.OAuthProviderTypeGoogle:
		return &GoogleImpl{
			URLPrefix:      p.urlPrefixProvider.Value(),
			OAuthConfig:    p.tenantConfig.UserConfig.SSO.OAuth,
			ProviderConfig: providerConfig,
		}
	case config.OAuthProviderTypeFacebook:
		return &FacebookImpl{
			URLPrefix:      p.urlPrefixProvider.Value(),
			OAuthConfig:    p.tenantConfig.UserConfig.SSO.OAuth,
			ProviderConfig: providerConfig,
		}
	case config.OAuthProviderTypeInstagram:
		return &InstagramImpl{
			URLPrefix:      p.urlPrefixProvider.Value(),
			OAuthConfig:    p.tenantConfig.UserConfig.SSO.OAuth,
			ProviderConfig: providerConfig,
		}
	case config.OAuthProviderTypeLinkedIn:
		return &LinkedInImpl{
			URLPrefix:      p.urlPrefixProvider.Value(),
			OAuthConfig:    p.tenantConfig.UserConfig.SSO.OAuth,
			ProviderConfig: providerConfig,
		}
	case config.OAuthProviderTypeAzureADv2:
		return &Azureadv2Impl{
			URLPrefix:      p.urlPrefixProvider.Value(),
			OAuthConfig:    p.tenantConfig.UserConfig.SSO.OAuth,
			ProviderConfig: providerConfig,
		}
	}
	return nil
}

func (p *ProviderFactory) GetProviderConfig(id string) (config.OAuthProviderConfiguration, bool) {
	return p.tenantConfig.GetOAuthProviderByID(id)
}

func ValidateCallbackURL(allowedCallbackURLs []string, callbackURL string) (err error) {
	// The logic of this function must be in sync with the inline javascript implementation.
	if callbackURL == "" {
		err = fmt.Errorf("missing callback URL")
		return
	}

	for _, v := range allowedCallbackURLs {
		if strings.HasPrefix(callbackURL, v) {
			return nil
		}
	}

	err = fmt.Errorf("callback URL is not whitelisted")
	return
}

func GenerateOpenIDConnectNonce() string {
	nonce := rand.StringWithAlphabet(32, nonceAlphabet, rand.SecureRand)
	return nonce
}
