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

func (f *MockSSOProvider) GetAuthInfo(r OAuthAuthorizationResponse) (authInfo AuthInfo, err error) {
	state, err := DecodeState(f.OAuthConfig.StateJWTSecret, r.State)
	if err != nil {
		return
	}

	authInfo = AuthInfo{
		ProviderConfig:          f.ProviderConfig,
		State:                   *state,
		ProviderAccessTokenResp: map[string]interface{}{},
		ProviderRawProfile:      map[string]interface{}{},
		ProviderUserInfo:        f.UserInfo,
	}
	return
}

func (f *MockSSOProvider) NonOpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse) (authInfo AuthInfo, err error) {
	return f.GetAuthInfo(r)
}

func (f *MockSSOProvider) OpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse) (authInfo AuthInfo, err error) {
	return f.GetAuthInfo(r)
}

func (f *MockSSOProvider) DecodeState(encodedState string) (*State, error) {
	return DecodeState(f.OAuthConfig.StateJWTSecret, encodedState)
}

func (f *MockSSOProvider) ExternalAccessTokenGetAuthInfo(accessTokenResp AccessTokenResp, state State) (authInfo AuthInfo, err error) {
	authInfo = AuthInfo{
		ProviderConfig:          f.ProviderConfig,
		ProviderAccessTokenResp: map[string]interface{}{},
		ProviderRawProfile:      map[string]interface{}{},
		ProviderUserInfo:        f.UserInfo,
	}
	return
}

var (
	_ OAuthProvider                   = &MockSSOProvider{}
	_ NonOpenIDConnectProvider        = &MockSSOProvider{}
	_ OpenIDConnectProvider           = &MockSSOProvider{}
	_ ExternalAccessTokenFlowProvider = &MockSSOProvider{}
)
