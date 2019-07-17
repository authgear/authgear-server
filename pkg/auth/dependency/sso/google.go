package sso

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
)

const (
	googleAuthorizationURL string = "https://accounts.google.com/o/oauth2/v2/auth"
	// nolint: gosec
	googleTokenURL    string = "https://www.googleapis.com/oauth2/v4/token"
	googleUserInfoURL string = "https://www.googleapis.com/oauth2/v1/userinfo"
)

type GoogleImpl struct {
	OAuthConfig    config.OAuthConfiguration
	ProviderConfig config.OAuthProviderConfiguration
}

func (f *GoogleImpl) GetAuthURL(params GetURLParams) (string, error) {
	params.Options["access_type"] = "offline"
	params.Options["prompt"] = "select_account"
	p := authURLParams{
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		options:        params.Options,
		state:          NewState(params),
		baseURL:        googleAuthorizationURL,
	}
	return authURL(p)
}

func (f *GoogleImpl) DecodeState(encodedState string) (*State, error) {
	return DecodeState(f.OAuthConfig.StateJWTSecret, encodedState)
}

func (f *GoogleImpl) GetAuthInfo(r OAuthAuthorizationResponse) (authInfo AuthInfo, err error) {
	return f.NonOpenIDConnectGetAuthInfo(r)
}

func (f *GoogleImpl) NonOpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse) (authInfo AuthInfo, err error) {
	h := getAuthInfoRequest{
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		code:           r.Code,
		encodedState:   r.State,
		accessTokenURL: googleTokenURL,
		userProfileURL: googleUserInfoURL,
		processor:      newDefaultAuthInfoProcessor(),
	}
	return h.getAuthInfo()
}

func (f *GoogleImpl) ExternalAccessTokenGetAuthInfo(accessTokenResp AccessTokenResp, state State) (authInfo AuthInfo, err error) {
	encodedState, err := EncodeState(f.OAuthConfig.StateJWTSecret, state)
	if err != nil {
		return
	}
	h := getAuthInfoRequest{
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		encodedState:   encodedState,
		accessTokenURL: googleTokenURL,
		userProfileURL: googleUserInfoURL,
		processor:      newDefaultAuthInfoProcessor(),
	}
	return h.getAuthInfoByAccessTokenResp(accessTokenResp)
}

var (
	_ OAuthProvider                   = &GoogleImpl{}
	_ NonOpenIDConnectProvider        = &GoogleImpl{}
	_ ExternalAccessTokenFlowProvider = &GoogleImpl{}
)
