package sso

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/errors"
)

type MockSSOProvider struct {
	BaseURL        string
	OAuthConfig    *config.OAuthConfiguration
	URLPrefix      *url.URL
	ProviderConfig config.OAuthProviderConfiguration
	UserInfo       ProviderUserInfo
}

func (f *MockSSOProvider) GetAuthURL(params GetURLParams) (string, error) {
	if f.ProviderConfig.ClientID == "" {
		return "", errors.New("must provide ClientID")
	}
	p := authURLParams{
		oauthConfig:    f.OAuthConfig,
		urlPrefix:      f.URLPrefix,
		providerConfig: f.ProviderConfig,
		state:          NewState(params),
		baseURL:        f.BaseURL,
	}
	return authURL(p)
}

func (f *MockSSOProvider) GetAuthInfo(r OAuthAuthorizationResponse) (authInfo AuthInfo, err error) {
	rawProfile := map[string]interface{}{
		"id": f.UserInfo.ID,
	}
	if f.UserInfo.Email != "" {
		rawProfile["email"] = f.UserInfo.Email
	}
	authInfo = AuthInfo{
		ProviderConfig:          f.ProviderConfig,
		ProviderAccessTokenResp: map[string]interface{}{},
		ProviderRawProfile:      rawProfile,
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

func (f *MockSSOProvider) EncodeState(state State) (encodedState string, err error) {
	return EncodeState(f.OAuthConfig.StateJWTSecret, state)
}

func (f *MockSSOProvider) DecodeState(encodedState string) (*State, error) {
	return DecodeState(f.OAuthConfig.StateJWTSecret, encodedState)
}

func (f *MockSSOProvider) EncodeSkygearAuthorizationCode(code SkygearAuthorizationCode) (encoded string, err error) {
	return EncodeSkygearAuthorizationCode(f.OAuthConfig.StateJWTSecret, code)
}

func (f *MockSSOProvider) DecodeSkygearAuthorizationCode(encoded string) (*SkygearAuthorizationCode, error) {
	return DecodeSkygearAuthorizationCode(f.OAuthConfig.StateJWTSecret, encoded)
}

func (f *MockSSOProvider) ExternalAccessTokenGetAuthInfo(accessTokenResp AccessTokenResp) (authInfo AuthInfo, err error) {
	rawProfile := map[string]interface{}{
		"id": f.UserInfo.ID,
	}
	if f.UserInfo.Email != "" {
		rawProfile["email"] = f.UserInfo.Email
	}
	authInfo = AuthInfo{
		ProviderConfig:          f.ProviderConfig,
		ProviderAccessTokenResp: map[string]interface{}{},
		ProviderRawProfile:      rawProfile,
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
