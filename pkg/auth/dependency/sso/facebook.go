package sso

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
)

const (
	facebookAuthorizationURL string = "https://www.facebook.com/v6.0/dialog/oauth"
	// nolint: gosec
	facebookTokenURL    string = "https://graph.facebook.com/v6.0/oauth/access_token"
	facebookUserInfoURL string = "https://graph.facebook.com/v6.0/me?fields=id,email"
)

type FacebookImpl struct {
	RedirectURL     RedirectURLProvider
	ProviderConfig  config.OAuthSSOProviderConfig
	Credentials     config.OAuthClientCredentialsItem
	UserInfoDecoder UserInfoDecoder
}

func (f *FacebookImpl) Type() config.OAuthSSOProviderType {
	return config.OAuthSSOProviderTypeFacebook
}

func (f *FacebookImpl) GetAuthURL(param GetAuthURLParam) (string, error) {
	p := authURLParams{
		redirectURI: f.RedirectURL.SSOCallbackURL(f.ProviderConfig).String(),
		clientID:    f.ProviderConfig.ClientID,
		scope:       f.ProviderConfig.Type.Scope(),
		state:       param.State,
		baseURL:     facebookAuthorizationURL,
	}
	return authURL(p)
}

func (f *FacebookImpl) GetAuthInfo(r OAuthAuthorizationResponse, param GetAuthInfoParam) (authInfo AuthInfo, err error) {
	return f.NonOpenIDConnectGetAuthInfo(r, param)
}

func (f *FacebookImpl) NonOpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, _ GetAuthInfoParam) (authInfo AuthInfo, err error) {
	h := getAuthInfoRequest{
		redirectURL:     f.RedirectURL.SSOCallbackURL(f.ProviderConfig).String(),
		providerConfig:  f.ProviderConfig,
		clientSecret:    f.Credentials.ClientSecret,
		accessTokenURL:  facebookTokenURL,
		userProfileURL:  facebookUserInfoURL,
		userInfoDecoder: f.UserInfoDecoder,
	}
	return h.getAuthInfo(r)
}

var (
	_ OAuthProvider            = &FacebookImpl{}
	_ NonOpenIDConnectProvider = &FacebookImpl{}
)
