package sso

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
)

const (
	linkedinAuthorizationURL string = "https://www.linkedin.com/oauth/v2/authorization"
	// nolint: gosec
	linkedinTokenURL    string = "https://www.linkedin.com/oauth/v2/accessToken"
	linkedinUserInfoURL string = "https://www.linkedin.com/v1/people/~?format=json"
)

type LinkedInImpl struct {
	OAuthConfig    config.OAuthConfiguration
	ProviderConfig config.OAuthProviderConfiguration
}

func (f *LinkedInImpl) GetAuthURL(params GetURLParams) (string, error) {
	p := authURLParams{
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		options:        params.Options,
		state:          NewState(params),
		baseURL:        linkedinAuthorizationURL,
	}
	return authURL(p)
}

func (f *LinkedInImpl) DecodeState(encodedState string) (*State, error) {
	return DecodeState(f.OAuthConfig.StateJWTSecret, encodedState)
}

func (f *LinkedInImpl) GetAuthInfo(code string, scope string, encodedState string) (authInfo AuthInfo, err error) {
	h := getAuthInfoRequest{
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		code:           code,
		encodedState:   encodedState,
		accessTokenURL: linkedinTokenURL,
		userProfileURL: linkedinUserInfoURL,
		processor:      newDefaultAuthInfoProcessor(),
	}
	return h.getAuthInfo()
}

func (f *LinkedInImpl) GetAuthInfoByAccessTokenResp(accessTokenResp AccessTokenResp) (authInfo AuthInfo, err error) {
	h := getAuthInfoRequest{
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		accessTokenURL: linkedinTokenURL,
		userProfileURL: linkedinUserInfoURL,
		processor:      newDefaultAuthInfoProcessor(),
	}
	return h.getAuthInfoByAccessTokenResp(accessTokenResp)
}

var (
	_ Provider = &LinkedInImpl{}
)
