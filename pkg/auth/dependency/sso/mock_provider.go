package sso

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"net/url"

	"github.com/skygeario/skygear-server/pkg/auth/model"
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

func (f *MockSSOProvider) Type() config.OAuthProviderType {
	return config.OAuthProviderTypeGoogle
}

func (f *MockSSOProvider) GetAuthURL(state State, encodedState string) (string, error) {
	if f.ProviderConfig.ClientID == "" {
		return "", errors.New("must provide ClientID")
	}
	p := authURLParams{
		oauthConfig:    f.OAuthConfig,
		urlPrefix:      f.URLPrefix,
		providerConfig: f.ProviderConfig,
		encodedState:   encodedState,
		baseURL:        f.BaseURL,
	}
	return authURL(p)
}

func (f *MockSSOProvider) GetAuthInfo(r OAuthAuthorizationResponse, state State) (authInfo AuthInfo, err error) {
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

func (f *MockSSOProvider) NonOpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, state State) (authInfo AuthInfo, err error) {
	return f.GetAuthInfo(r, state)
}

func (f *MockSSOProvider) OpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, state State) (authInfo AuthInfo, err error) {
	return f.GetAuthInfo(r, state)
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

func (f *MockSSOProvider) EncodeState(state State) (encodedState string, err error) {
	return EncodeState(f.OAuthConfig.StateJWTSecret, "myapp", state)
}

func (f *MockSSOProvider) DecodeState(encodedState string) (*State, error) {
	return DecodeState(f.OAuthConfig.StateJWTSecret, "myapp", encodedState)
}

func (f *MockSSOProvider) EncodeSkygearAuthorizationCode(code SkygearAuthorizationCode) (encoded string, err error) {
	return EncodeSkygearAuthorizationCode(f.OAuthConfig.StateJWTSecret, "myapp", code)
}

func (f *MockSSOProvider) DecodeSkygearAuthorizationCode(encoded string) (*SkygearAuthorizationCode, error) {
	return DecodeSkygearAuthorizationCode(f.OAuthConfig.StateJWTSecret, "myapp", encoded)
}

func (f *MockSSOProvider) IsAllowedOnUserDuplicate(a model.OnUserDuplicate) bool {
	return model.IsAllowedOnUserDuplicate(
		f.OAuthConfig.OnUserDuplicateAllowMerge,
		f.OAuthConfig.OnUserDuplicateAllowCreate,
		a,
	)
}

func (f *MockSSOProvider) IsValidCallbackURL(client config.OAuthClientConfiguration, u string) bool {
	err := ValidateCallbackURL(client.RedirectURIs(), u)
	return err == nil
}

func (f *MockSSOProvider) IsExternalAccessTokenFlowEnabled() bool {
	return f.OAuthConfig.ExternalAccessTokenFlowEnabled
}

func (f *MockSSOProvider) VerifyPKCE(code *SkygearAuthorizationCode, codeVerifier string) error {
	sha256Arr := sha256.Sum256([]byte(codeVerifier))
	sha256Slice := sha256Arr[:]
	codeChallenge := base64.RawURLEncoding.EncodeToString(sha256Slice)
	if subtle.ConstantTimeCompare([]byte(code.CodeChallenge), []byte(codeChallenge)) != 1 {
		return NewSSOFailed(InvalidCodeVerifier, "invalid code verifier")
	}
	return nil
}

var (
	_ OAuthProvider                   = &MockSSOProvider{}
	_ NonOpenIDConnectProvider        = &MockSSOProvider{}
	_ OpenIDConnectProvider           = &MockSSOProvider{}
	_ ExternalAccessTokenFlowProvider = &MockSSOProvider{}
	_ Provider                        = &MockSSOProvider{}
)
