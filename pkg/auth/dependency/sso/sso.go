package sso

import (
	"strings"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

// Scope parameter allows the application to express the desired scope of the access request.
type Scope []string

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
	Scope       Scope
	Options     Options
	CallbackURL string
	UXMode      UXMode
	UserID      string
	Action      string
}

// Setting is the base settings for SSO
type Setting struct {
	URLPrefix           string
	JSSDKCDNURL         string
	StateJWTSecret      string
	AutoLinkEnabled     bool
	AllowedCallbackURLs []string
}

// Config is the base config of a SSO provider
type Config struct {
	Name         string
	ClientID     string
	ClientSecret string
	Scope        Scope
}

// AuthInfo contains auth info from HandleAuthzResp
type AuthInfo struct {
	ProviderName            string
	ProviderUserID          string
	ProviderUserProfile     map[string]interface{}
	ProviderAccessTokenResp interface{}
	ProviderAuthData        ProviderAuthKeys
	State                   State
}

type ProviderAuthKeys struct {
	Email string
}

// Provider defines SSO interface
type Provider interface {
	GetAuthURL(params GetURLParams) (url string, err error)
	GetAuthInfo(code string, scope Scope, encodedState string) (authInfo AuthInfo, err error)
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

func (p *ProviderFactory) NewProvider(name string) Provider {
	SSOConf, ok := p.tenantConfig.GetSSOProviderByName(name)
	if !ok {
		return nil
	}
	SSOSetting := p.tenantConfig.UserConfig.SSO
	setting := Setting{
		URLPrefix:           SSOSetting.URLPrefix,
		JSSDKCDNURL:         SSOSetting.JSSDKCDNURL,
		StateJWTSecret:      SSOSetting.StateJWTSecret,
		AutoLinkEnabled:     SSOSetting.AutoLinkEnabled,
		AllowedCallbackURLs: SSOSetting.AllowedCallbackURLs,
	}
	config := Config{
		Name:         SSOConf.Name,
		ClientID:     SSOConf.ClientID,
		ClientSecret: SSOConf.ClientSecret,
		Scope:        strings.Split(SSOConf.Scope, ","),
	}

	switch name {
	case "google":
		return &GoogleImpl{
			Setting: setting,
			Config:  config,
		}
	case "facebook":
		return &FacebookImpl{
			Setting: setting,
			Config:  config,
		}
	case "instagram":
		return &InstagramImpl{
			Setting: setting,
			Config:  config,
		}
	case "linkedin":
		return &LinkedInImpl{
			Setting: setting,
			Config:  config,
		}
	}
	return nil
}

func (p *ProviderFactory) Setting() Setting {
	SSOSetting := p.tenantConfig.UserConfig.SSO
	return Setting{
		URLPrefix:           SSOSetting.URLPrefix,
		JSSDKCDNURL:         SSOSetting.JSSDKCDNURL,
		StateJWTSecret:      SSOSetting.StateJWTSecret,
		AutoLinkEnabled:     SSOSetting.AutoLinkEnabled,
		AllowedCallbackURLs: SSOSetting.AllowedCallbackURLs,
	}
}
