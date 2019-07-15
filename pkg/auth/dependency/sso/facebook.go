package sso

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
)

const (
	facebookAuthorizationURL string = "https://www.facebook.com/dialog/oauth"
	// nolint: gosec
	facebookTokenURL    string = "https://graph.facebook.com/v2.10/oauth/access_token"
	facebookUserInfoURL string = "https://graph.facebook.com/v2.10/me"
)

type FacebookImpl struct {
	OAuthConfig    config.OAuthConfiguration
	ProviderConfig config.OAuthProviderConfiguration
}

func (f *FacebookImpl) GetAuthURL(params GetURLParams) (string, error) {
	if params.State.UXMode == UXModeWebPopup {
		// https://developers.facebook.com/docs/facebook-login/manually-build-a-login-flow
		params.Options["display"] = "popup"
	}
	p := authURLParams{
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		options:        params.Options,
		state:          NewState(params),
		baseURL:        facebookAuthorizationURL,
	}
	return authURL(p)
}

func (f *FacebookImpl) GetAuthInfo(code string, scope string, encodedState string) (authInfo AuthInfo, err error) {
	h := getAuthInfoRequest{
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		code:           code,
		encodedState:   encodedState,
		accessTokenURL: facebookTokenURL,
		userProfileURL: facebookUserInfoURL,
		processor:      newDefaultAuthInfoProcessor(),
	}
	return h.getAuthInfo()
}

func (f *FacebookImpl) DecodeState(encodedState string) (*State, error) {
	return DecodeState(f.OAuthConfig.StateJWTSecret, encodedState)
}

func (f *FacebookImpl) GetAuthInfoByAccessTokenResp(accessTokenResp AccessTokenResp) (authInfo AuthInfo, err error) {
	h := getAuthInfoRequest{
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		accessTokenURL: facebookTokenURL,
		userProfileURL: facebookUserInfoURL,
		processor:      newDefaultAuthInfoProcessor(),
	}
	return h.getAuthInfoByAccessTokenResp(accessTokenResp)
}

var (
	_ Provider = &FacebookImpl{}
)
