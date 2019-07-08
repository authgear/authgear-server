package sso

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
)

// Options parameter allows additional options for getting auth url
type Options map[string]interface{}

// State parameter refers parameter of auth url
type State struct {
	UXMode      string `json:"ux_mode"`
	CallbackURL string `json:"callback_url"`
	Action      string `json:"action"`
	UserID      string `json:"user_id,omitempty"`
}

// UXMode indicates how the URL is used
type UXMode int

const (
	// Undefined for undefined uxmode
	Undefined UXMode = iota
	// WebRedirect for web url redirect
	WebRedirect
	// WebPopup for web popup window
	WebPopup
	// IOS for device iOS
	IOS
	// Android for device Android
	Android
)

func (m UXMode) String() string {
	names := [...]string{
		"web_redirect",
		"web_popup",
		"ios",
		"android",
	}

	if m < WebRedirect || m > Android {
		return "undefined"
	}

	return names[m-1]
}

// UXModeFromString converts string to UXMode
func UXModeFromString(input string) (u UXMode) {
	UXModes := [...]UXMode{WebRedirect, WebPopup, IOS, Android}
	for _, v := range UXModes {
		if input == v.String() {
			u = v
			return
		}
	}

	return
}

// GetURLParams structs parameters for GetLoginAuthURL
type GetURLParams struct {
	Options     Options
	CallbackURL string
	UXMode      UXMode
	UserID      string
	Action      string
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
