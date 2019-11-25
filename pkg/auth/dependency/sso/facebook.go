package sso

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

const (
	facebookAuthorizationURL string = "https://www.facebook.com/dialog/oauth"
	// nolint: gosec
	facebookTokenURL    string = "https://graph.facebook.com/v2.10/oauth/access_token"
	facebookUserInfoURL string = "https://graph.facebook.com/v2.10/me"
)

type FacebookImpl struct {
	URLPrefix      *url.URL
	OAuthConfig    *config.OAuthConfiguration
	ProviderConfig config.OAuthProviderConfiguration
}

func (f *FacebookImpl) GetAuthURL(params GetURLParams) (string, error) {
	p := authURLParams{
		oauthConfig:    f.OAuthConfig,
		urlPrefix:      f.URLPrefix,
		providerConfig: f.ProviderConfig,
		state:          NewState(params),
		baseURL:        facebookAuthorizationURL,
	}
	if params.State.UXMode == UXModeWebPopup {
		// https://developers.facebook.com/docs/facebook-login/manually-build-a-login-flow
		p.display = "popup"
	}
	return authURL(p)
}

func (f *FacebookImpl) GetAuthInfo(r OAuthAuthorizationResponse) (authInfo AuthInfo, err error) {
	return f.NonOpenIDConnectGetAuthInfo(r)
}

func (f *FacebookImpl) NonOpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse) (authInfo AuthInfo, err error) {
	h := getAuthInfoRequest{
		urlPrefix:      f.URLPrefix,
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		accessTokenURL: facebookTokenURL,
		userProfileURL: facebookUserInfoURL,
		processor:      NewDefaultUserInfoDecoder(),
	}
	return h.getAuthInfo(r)
}

func (f *FacebookImpl) EncodeState(state State) (encodedState string, err error) {
	return EncodeState(f.OAuthConfig.StateJWTSecret, state)
}

func (f *FacebookImpl) DecodeState(encodedState string) (*State, error) {
	return DecodeState(f.OAuthConfig.StateJWTSecret, encodedState)
}

func (f *FacebookImpl) ExternalAccessTokenGetAuthInfo(accessTokenResp AccessTokenResp) (authInfo AuthInfo, err error) {
	h := getAuthInfoRequest{
		urlPrefix:      f.URLPrefix,
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		accessTokenURL: facebookTokenURL,
		userProfileURL: facebookUserInfoURL,
		processor:      NewDefaultUserInfoDecoder(),
	}
	return h.getAuthInfoByAccessTokenResp(accessTokenResp)
}

var (
	_ OAuthProvider                   = &FacebookImpl{}
	_ NonOpenIDConnectProvider        = &FacebookImpl{}
	_ ExternalAccessTokenFlowProvider = &FacebookImpl{}
)
