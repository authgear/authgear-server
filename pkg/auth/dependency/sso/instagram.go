package sso

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

const (
	instagramAuthorizationURL string = "https://api.instagram.com/oauth/authorize"
	// nolint: gosec
	instagramTokenURL    string = "https://api.instagram.com/oauth/access_token"
	instagramUserInfoURL string = "https://api.instagram.com/v1/users/self"
)

type InstagramImpl struct {
	URLPrefix      *url.URL
	OAuthConfig    *config.OAuthConfiguration
	ProviderConfig config.OAuthProviderConfiguration
}

func (f *InstagramImpl) GetAuthURL(state State, encodedState string) (string, error) {
	p := authURLParams{
		oauthConfig:    f.OAuthConfig,
		urlPrefix:      f.URLPrefix,
		providerConfig: f.ProviderConfig,
		encodedState:   encodedState,
		baseURL:        instagramAuthorizationURL,
	}
	return authURL(p)
}

func (f *InstagramImpl) GetAuthInfo(r OAuthAuthorizationResponse, state State) (authInfo AuthInfo, err error) {
	return f.NonOpenIDConnectGetAuthInfo(r, state)
}

func (f *InstagramImpl) NonOpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, state State) (authInfo AuthInfo, err error) {
	h := getAuthInfoRequest{
		urlPrefix:      f.URLPrefix,
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		accessTokenURL: instagramTokenURL,
		userProfileURL: instagramUserInfoURL,
		processor:      NewInstagramUserInfoDecoder(),
	}
	return h.getAuthInfo(r, state)
}
func (f *InstagramImpl) ExternalAccessTokenGetAuthInfo(accessTokenResp AccessTokenResp) (authInfo AuthInfo, err error) {
	h := getAuthInfoRequest{
		urlPrefix:      f.URLPrefix,
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		accessTokenURL: instagramTokenURL,
		userProfileURL: instagramUserInfoURL,
		processor:      NewInstagramUserInfoDecoder(),
	}
	return h.getAuthInfoByAccessTokenResp(accessTokenResp)
}

var (
	_ OAuthProvider                   = &InstagramImpl{}
	_ NonOpenIDConnectProvider        = &InstagramImpl{}
	_ ExternalAccessTokenFlowProvider = &InstagramImpl{}
)
