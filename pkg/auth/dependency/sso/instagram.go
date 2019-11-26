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

func (f *InstagramImpl) GetAuthURL(state State) (string, error) {
	p := authURLParams{
		oauthConfig:    f.OAuthConfig,
		urlPrefix:      f.URLPrefix,
		providerConfig: f.ProviderConfig,
		state:          state,
		baseURL:        instagramAuthorizationURL,
	}
	return authURL(p)
}

func (f *InstagramImpl) EncodeState(state State) (encodedState string, err error) {
	return EncodeState(f.OAuthConfig.StateJWTSecret, state)
}

func (f *InstagramImpl) DecodeState(encodedState string) (*State, error) {
	return DecodeState(f.OAuthConfig.StateJWTSecret, encodedState)
}

func (f *InstagramImpl) EncodeSkygearAuthorizationCode(code SkygearAuthorizationCode) (encoded string, err error) {
	return EncodeSkygearAuthorizationCode(f.OAuthConfig.StateJWTSecret, code)
}

func (f *InstagramImpl) DecodeSkygearAuthorizationCode(encoded string) (*SkygearAuthorizationCode, error) {
	return DecodeSkygearAuthorizationCode(f.OAuthConfig.StateJWTSecret, encoded)
}

func (f *InstagramImpl) GetAuthInfo(r OAuthAuthorizationResponse) (authInfo AuthInfo, err error) {
	return f.NonOpenIDConnectGetAuthInfo(r)
}

func (f *InstagramImpl) NonOpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse) (authInfo AuthInfo, err error) {
	h := getAuthInfoRequest{
		urlPrefix:      f.URLPrefix,
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		accessTokenURL: instagramTokenURL,
		userProfileURL: instagramUserInfoURL,
		processor:      NewInstagramUserInfoDecoder(),
	}
	return h.getAuthInfo(r)
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
