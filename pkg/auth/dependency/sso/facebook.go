package sso

import (
	"io"
	"strings"

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

type facebookAuthInfoProcessor struct {
	defaultAuthInfoProcessor
}

func newFacebookAuthInfoProcessor() facebookAuthInfoProcessor {
	return facebookAuthInfoProcessor{}
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
	p := newFacebookAuthInfoProcessor()
	h := getAuthInfoRequest{
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		code:           code,
		encodedState:   encodedState,
		accessTokenURL: facebookTokenURL,
		userProfileURL: facebookUserInfoURL,
		processor:      p,
	}
	return h.getAuthInfo()
}

func (f *FacebookImpl) DecodeState(encodedState string) (*State, error) {
	return DecodeState(f.OAuthConfig.StateJWTSecret, encodedState)
}

func (f facebookAuthInfoProcessor) DecodeAccessTokenResp(r io.Reader) (AccessTokenResp, error) {
	accessTokenResp, err := f.defaultAuthInfoProcessor.DecodeAccessTokenResp(r)
	if err != nil {
		return accessTokenResp, err
	}

	// special handling for facebook access token
	if accessTokenResp.ExpiresIn == 0 && accessTokenResp.RawExpires != 0 {
		accessTokenResp.ExpiresIn = accessTokenResp.RawExpires
	}
	if strings.ToLower(accessTokenResp.TokenType) == "bearer" {
		accessTokenResp.TokenType = "Bearer"
	}
	return accessTokenResp, nil
}

func (f *FacebookImpl) GetAuthInfoByAccessTokenResp(accessTokenResp AccessTokenResp) (authInfo AuthInfo, err error) {
	p := newFacebookAuthInfoProcessor()
	h := getAuthInfoRequest{
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		accessTokenURL: facebookTokenURL,
		userProfileURL: facebookUserInfoURL,
		processor:      p,
	}
	return h.getAuthInfoByAccessTokenResp(accessTokenResp)
}

var (
	_ Provider = &FacebookImpl{}
)
