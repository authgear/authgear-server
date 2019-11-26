package sso

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

const (
	linkedinAuthorizationURL string = "https://www.linkedin.com/oauth/v2/authorization"
	// nolint: gosec
	linkedinTokenURL    string = "https://www.linkedin.com/oauth/v2/accessToken"
	linkedinUserInfoURL string = "https://www.linkedin.com/v1/people/~?format=json"
)

type LinkedInImpl struct {
	URLPrefix      *url.URL
	OAuthConfig    *config.OAuthConfiguration
	ProviderConfig config.OAuthProviderConfiguration
}

func (f *LinkedInImpl) GetAuthURL(state State) (string, error) {
	p := authURLParams{
		oauthConfig:    f.OAuthConfig,
		urlPrefix:      f.URLPrefix,
		providerConfig: f.ProviderConfig,
		state:          state,
		baseURL:        linkedinAuthorizationURL,
	}
	return authURL(p)
}

func (f *LinkedInImpl) EncodeState(state State) (encodedState string, err error) {
	return EncodeState(f.OAuthConfig.StateJWTSecret, state)
}

func (f *LinkedInImpl) DecodeState(encodedState string) (*State, error) {
	return DecodeState(f.OAuthConfig.StateJWTSecret, encodedState)
}

func (f *LinkedInImpl) EncodeSkygearAuthorizationCode(code SkygearAuthorizationCode) (encoded string, err error) {
	return EncodeSkygearAuthorizationCode(f.OAuthConfig.StateJWTSecret, code)
}

func (f *LinkedInImpl) DecodeSkygearAuthorizationCode(encoded string) (*SkygearAuthorizationCode, error) {
	return DecodeSkygearAuthorizationCode(f.OAuthConfig.StateJWTSecret, encoded)
}

func (f *LinkedInImpl) GetAuthInfo(r OAuthAuthorizationResponse) (authInfo AuthInfo, err error) {
	return f.NonOpenIDConnectGetAuthInfo(r)
}

func (f *LinkedInImpl) NonOpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse) (authInfo AuthInfo, err error) {
	h := getAuthInfoRequest{
		urlPrefix:      f.URLPrefix,
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		accessTokenURL: linkedinTokenURL,
		userProfileURL: linkedinUserInfoURL,
		processor:      NewDefaultUserInfoDecoder(),
	}
	return h.getAuthInfo(r)
}

func (f *LinkedInImpl) ExternalAccessTokenGetAuthInfo(accessTokenResp AccessTokenResp) (authInfo AuthInfo, err error) {
	h := getAuthInfoRequest{
		urlPrefix:      f.URLPrefix,
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		accessTokenURL: linkedinTokenURL,
		userProfileURL: linkedinUserInfoURL,
		processor:      NewDefaultUserInfoDecoder(),
	}
	return h.getAuthInfoByAccessTokenResp(accessTokenResp)
}

var (
	_ OAuthProvider                   = &LinkedInImpl{}
	_ NonOpenIDConnectProvider        = &LinkedInImpl{}
	_ ExternalAccessTokenFlowProvider = &LinkedInImpl{}
)
