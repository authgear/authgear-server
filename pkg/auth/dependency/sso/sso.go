package sso

import (
	"fmt"
	"strings"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/rand"
)

const (
	nonceAlphabet string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// Options parameter allows additional options for getting auth url
type Options map[string]interface{}

// State is an opaque value used by the client to maintain
// state between the request and callback.
// See https://tools.ietf.org/html/rfc6749#section-4.1.1
type State struct {
	UXMode          UXMode          `json:"ux_mode"`
	CallbackURL     string          `json:"callback_url"`
	Action          string          `json:"action"`
	UserID          string          `json:"user_id,omitempty"`
	MergeRealm      string          `json:"merge_realm,omitempty"`
	OnUserDuplicate OnUserDuplicate `json:"on_user_duplicate,omitempty"`
}

// UXMode indicates how the URL is used
type UXMode string

// UXMode constants
const (
	UXModeWebRedirect UXMode = "web_redirect"
	UXModeWebPopup    UXMode = "web_popup"
	UXModeIOS         UXMode = "ios"
	UXModeAndroid     UXMode = "android"
)

// IsValidUXMode validates UXMode
func IsValidUXMode(mode UXMode) bool {
	allModes := []UXMode{UXModeWebRedirect, UXModeWebPopup, UXModeIOS, UXModeAndroid}
	for _, v := range allModes {
		if mode == v {
			return true
		}
	}
	return false
}

// OnUserDuplicate is the strategy to handle user duplicate
type OnUserDuplicate string

// OnUserDuplicate constants
const (
	OnUserDuplicateAbort  OnUserDuplicate = "abort"
	OnUserDuplicateMerge  OnUserDuplicate = "merge"
	OnUserDuplicateCreate OnUserDuplicate = "create"
)

// OnUserDuplicateDefault is OnUserDuplicateAbort
const OnUserDuplicateDefault = OnUserDuplicateAbort

// IsValidOnUserDuplicate validates OnUserDuplicate
func IsValidOnUserDuplicate(input OnUserDuplicate) bool {
	allVariants := []OnUserDuplicate{OnUserDuplicateAbort, OnUserDuplicateMerge, OnUserDuplicateCreate}
	for _, v := range allVariants {
		if input == v {
			return true
		}
	}
	return false
}

// IsAllowedOnUserDuplicate checks if input is allowed
func IsAllowedOnUserDuplicate(onUserDuplicateAllowMerge bool, onUserDuplicateAllowCreate bool, input OnUserDuplicate) bool {
	if !onUserDuplicateAllowMerge && input == OnUserDuplicateMerge {
		return false
	}
	if !onUserDuplicateAllowCreate && input == OnUserDuplicateCreate {
		return false
	}
	return true
}

// GetURLParams is the argument of getAuthURL
type GetURLParams struct {
	Options Options
	State   State
	Nonce   string
}

// AuthInfo contains auth info from HandleAuthzResp
type AuthInfo struct {
	ProviderConfig          config.OAuthProviderConfiguration
	ProviderRawProfile      map[string]interface{}
	ProviderAccessTokenResp interface{}
	ProviderUserInfo        ProviderUserInfo
	State                   State
}

type ProviderUserInfo struct {
	ID    string
	Email string
}

type OAuthAuthorizationResponse struct {
	Code  string
	State string
	Scope string
	// Nonce is required when the underlying provider is OpenID connect compliant.
	// The implementation is based on the suggestion in the spec.
	// See https://openid.net/specs/openid-connect-core-1_0.html#NonceNotes
	//
	// The nonce is a cryptographically random string.
	// The nonce passed to the provider is SHA256 hash of it.
	// The get-authorization URL endpoint ensures the nonce in session cookie.
	// The callback endpoint expect the user agent to include the nonce in session cookie.
	// The nonce in session cookie will be validated against the hashed nonce in the ID token.
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
	ExternalAccessTokenGetAuthInfo(accessTokenResp AccessTokenResp) (authInfo AuthInfo, err error)
}

// OpenIDConnectProvider are OpenID Connect provider.
// They are Azure AD v2.
type OpenIDConnectProvider interface {
	OpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse) (authInfo AuthInfo, err error)
}

type ProviderFactory struct {
	tenantConfig config.TenantConfiguration
}

func NewProviderFactory(tenantConfig config.TenantConfiguration) *ProviderFactory {
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
			OAuthConfig:    p.tenantConfig.UserConfig.SSO.OAuth,
			ProviderConfig: providerConfig,
		}
	case config.OAuthProviderTypeFacebook:
		return &FacebookImpl{
			OAuthConfig:    p.tenantConfig.UserConfig.SSO.OAuth,
			ProviderConfig: providerConfig,
		}
	case config.OAuthProviderTypeInstagram:
		return &InstagramImpl{
			OAuthConfig:    p.tenantConfig.UserConfig.SSO.OAuth,
			ProviderConfig: providerConfig,
		}
	case config.OAuthProviderTypeLinkedIn:
		return &LinkedInImpl{
			OAuthConfig:    p.tenantConfig.UserConfig.SSO.OAuth,
			ProviderConfig: providerConfig,
		}
	case config.OAuthProviderTypeAzureADv2:
		return &Azureadv2Impl{
			OAuthConfig:    p.tenantConfig.UserConfig.SSO.OAuth,
			ProviderConfig: providerConfig,
		}
	}
	return nil
}

func ValidateCallbackURL(allowedCallbackURLs []string, callbackURL string) (err error) {
	if callbackURL == "" {
		err = fmt.Errorf("missing callback URL")
		return
	}

	lowerCallbackURL := strings.ToLower(callbackURL)
	for _, v := range allowedCallbackURLs {
		lowerAllowed := strings.ToLower(v)
		if strings.HasPrefix(lowerCallbackURL, lowerAllowed) {
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
