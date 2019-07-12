package sso

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
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

// Provider defines SSO interface
type Provider interface {
	GetAuthURL(params GetURLParams) (url string, err error)
	// TODO: Remove scope
	GetAuthInfo(code string, scope string, encodedState string) (authInfo AuthInfo, err error)
	DecodeState(encodedState string) (*State, error)
	GetAuthInfoByAccessTokenResp(accessTokenResp AccessTokenResp) (authInfo AuthInfo, err error)
}

type ProviderFactory struct {
	tenantConfig config.TenantConfiguration
}

func NewProviderFactory(tenantConfig config.TenantConfiguration) *ProviderFactory {
	return &ProviderFactory{
		tenantConfig: tenantConfig,
	}
}

func (p *ProviderFactory) NewProvider(id string) Provider {
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
	}
	return nil
}
