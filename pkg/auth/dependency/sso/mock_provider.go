package sso

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

type MockSSOProvider struct {
	BaseURL        string
	OAuthConfig    config.OAuthConfiguration
	ProviderConfig config.OAuthProviderConfiguration
	UserInfo       ProviderUserInfo
}

func (f *MockSSOProvider) GetAuthURL(params GetURLParams) (string, error) {
	if f.ProviderConfig.ClientID == "" {
		skyErr := skyerr.NewError(skyerr.InvalidArgument, "ClientID is required")
		return "", skyErr
	}
	p := authURLParams{
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		options:        params.Options,
		state:          NewState(params),
		baseURL:        f.BaseURL,
	}
	return authURL(p)
}

func (f *MockSSOProvider) GetAuthInfo(code string, scope string, encodedState string) (authInfo AuthInfo, err error) {
	state, err := DecodeState(f.OAuthConfig.StateJWTSecret, encodedState)
	if err != nil {
		return
	}

	authInfo = AuthInfo{
		ProviderConfig:          f.ProviderConfig,
		State:                   state,
		ProviderAccessTokenResp: map[string]interface{}{},
		ProviderRawProfile:      map[string]interface{}{},
		ProviderUserInfo:        f.UserInfo,
	}
	return
}

func (f *MockSSOProvider) GetAuthInfoByAccessTokenResp(accessTokenResp AccessTokenResp) (authInfo AuthInfo, err error) {
	authInfo = AuthInfo{
		ProviderConfig:          f.ProviderConfig,
		ProviderAccessTokenResp: map[string]interface{}{},
		ProviderRawProfile:      map[string]interface{}{},
		ProviderUserInfo:        f.UserInfo,
	}
	return
}
