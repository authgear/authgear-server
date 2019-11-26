package sso

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

const (
	googleAuthorizationURL string = "https://accounts.google.com/o/oauth2/v2/auth"
	// nolint: gosec
	googleTokenURL    string = "https://www.googleapis.com/oauth2/v4/token"
	googleUserInfoURL string = "https://www.googleapis.com/oauth2/v1/userinfo"
)

type GoogleImpl struct {
	URLPrefix      *url.URL
	OAuthConfig    *config.OAuthConfiguration
	ProviderConfig config.OAuthProviderConfiguration
}

func (f *GoogleImpl) GetAuthURL(state State) (string, error) {
	p := authURLParams{
		oauthConfig:    f.OAuthConfig,
		urlPrefix:      f.URLPrefix,
		providerConfig: f.ProviderConfig,
		state:          state,
		baseURL:        googleAuthorizationURL,
		prompt:         "select_account",
	}
	return authURL(p)
}

func (f *GoogleImpl) GetAuthInfo(r OAuthAuthorizationResponse) (authInfo AuthInfo, err error) {
	return f.NonOpenIDConnectGetAuthInfo(r)
}

func (f *GoogleImpl) NonOpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse) (authInfo AuthInfo, err error) {
	h := getAuthInfoRequest{
		urlPrefix:      f.URLPrefix,
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		accessTokenURL: googleTokenURL,
		userProfileURL: googleUserInfoURL,
		processor:      NewDefaultUserInfoDecoder(),
	}
	return h.getAuthInfo(r)
}

func (f *GoogleImpl) ExternalAccessTokenGetAuthInfo(accessTokenResp AccessTokenResp) (authInfo AuthInfo, err error) {
	h := getAuthInfoRequest{
		urlPrefix:      f.URLPrefix,
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		accessTokenURL: googleTokenURL,
		userProfileURL: googleUserInfoURL,
		processor:      NewDefaultUserInfoDecoder(),
	}
	return h.getAuthInfoByAccessTokenResp(accessTokenResp)
}

var (
	_ OAuthProvider                   = &GoogleImpl{}
	_ NonOpenIDConnectProvider        = &GoogleImpl{}
	_ ExternalAccessTokenFlowProvider = &GoogleImpl{}
)
