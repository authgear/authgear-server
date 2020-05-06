package sso

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

const (
	linkedinAuthorizationURL string = "https://www.linkedin.com/oauth/v2/authorization"
	// nolint: gosec
	linkedinTokenURL    string = "https://www.linkedin.com/oauth/v2/accessToken"
	linkedinUserInfoURL string = "https://api.linkedin.com/v2/me"
)

type LinkedInImpl struct {
	URLPrefix       *url.URL
	RedirectURLFunc RedirectURLFunc
	OAuthConfig     *config.OAuthConfiguration
	ProviderConfig  config.OAuthProviderConfiguration
	UserInfoDecoder UserInfoDecoder
}

func (f *LinkedInImpl) Type() config.OAuthProviderType {
	return config.OAuthProviderTypeLinkedIn
}

func (f *LinkedInImpl) GetAuthURL(state State, encodedState string) (string, error) {
	p := authURLParams{
		oauthConfig:    f.OAuthConfig,
		redirectURI:    f.RedirectURLFunc(f.URLPrefix, f.ProviderConfig),
		providerConfig: f.ProviderConfig,
		encodedState:   encodedState,
		baseURL:        linkedinAuthorizationURL,
	}
	return authURL(p)
}

func (f *LinkedInImpl) GetAuthInfo(r OAuthAuthorizationResponse, state State) (authInfo AuthInfo, err error) {
	return f.NonOpenIDConnectGetAuthInfo(r, state)
}

func (f *LinkedInImpl) NonOpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, state State) (authInfo AuthInfo, err error) {
	h := getAuthInfoRequest{
		redirectURL:     f.RedirectURLFunc(f.URLPrefix, f.ProviderConfig),
		oauthConfig:     f.OAuthConfig,
		providerConfig:  f.ProviderConfig,
		accessTokenURL:  linkedinTokenURL,
		userProfileURL:  linkedinUserInfoURL,
		userInfoDecoder: f.UserInfoDecoder,
	}
	return h.getAuthInfo(r, state)
}

func (f *LinkedInImpl) ExternalAccessTokenGetAuthInfo(accessTokenResp AccessTokenResp) (authInfo AuthInfo, err error) {
	h := getAuthInfoRequest{
		redirectURL:     f.RedirectURLFunc(f.URLPrefix, f.ProviderConfig),
		oauthConfig:     f.OAuthConfig,
		providerConfig:  f.ProviderConfig,
		accessTokenURL:  linkedinTokenURL,
		userProfileURL:  linkedinUserInfoURL,
		userInfoDecoder: f.UserInfoDecoder,
	}
	return h.getAuthInfoByAccessTokenResp(accessTokenResp)
}

var (
	_ OAuthProvider                   = &LinkedInImpl{}
	_ NonOpenIDConnectProvider        = &LinkedInImpl{}
	_ ExternalAccessTokenFlowProvider = &LinkedInImpl{}
)
