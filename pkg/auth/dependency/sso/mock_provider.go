package sso

import (
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

type MockSSOProvider struct {
	BaseURL  string
	Setting  Setting
	Config   Config
	UserInfo ProviderUserInfo
}

func (f *MockSSOProvider) GetAuthURL(params GetURLParams) (string, error) {
	if f.Config.ClientID == "" {
		skyErr := skyerr.NewError(skyerr.InvalidArgument, "ClientID is required")
		return "", skyErr
	}
	p := authURLParams{
		providerName:   f.Config.Name,
		clientID:       f.Config.ClientID,
		urlPrefix:      f.Setting.URLPrefix,
		scope:          GetScope(params.Scope, f.Config.Scope),
		options:        params.Options,
		stateJWTSecret: f.Setting.StateJWTSecret,
		state:          NewState(params),
		baseURL:        f.BaseURL,
	}
	return authURL(p)
}

func (f *MockSSOProvider) GetAuthInfo(code string, scope Scope, encodedState string) (authInfo AuthInfo, err error) {
	state, err := DecodeState(f.Setting.StateJWTSecret, encodedState)
	if err != nil {
		return
	}

	authInfo = AuthInfo{
		ProviderName:            f.Config.Name,
		State:                   state,
		ProviderAccessTokenResp: map[string]interface{}{},
		ProviderRawProfile:      map[string]interface{}{},
		ProviderUserInfo:        f.UserInfo,
	}
	return
}

func (f *MockSSOProvider) GetAuthInfoByAccessTokenResp(accessTokenResp AccessTokenResp) (authInfo AuthInfo, err error) {
	authInfo = AuthInfo{
		ProviderName:            f.Config.Name,
		ProviderAccessTokenResp: map[string]interface{}{},
		ProviderRawProfile:      map[string]interface{}{},
		ProviderUserInfo:        f.UserInfo,
	}
	return
}
