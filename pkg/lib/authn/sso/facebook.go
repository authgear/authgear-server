package sso

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/crypto"
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

func (*FacebookImpl) Type() config.OAuthSSOProviderType {
	return config.OAuthSSOProviderTypeFacebook
}

func (f *FacebookImpl) Config() config.OAuthSSOProviderConfig {
	return f.ProviderConfig
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
	authInfo = AuthInfo{
		ProviderConfig: f.ProviderConfig,
	}

	accessTokenResp, err := fetchAccessTokenResp(
		r.Code,
		facebookTokenURL,
		f.RedirectURL.SSOCallbackURL(f.ProviderConfig).String(),
		f.ProviderConfig.ClientID,
		f.Credentials.ClientSecret,
	)
	if err != nil {
		return
	}
	authInfo.ProviderAccessTokenResp = accessTokenResp

	userProfileURL, err := url.Parse(facebookUserInfoURL)
	if err != nil {
		return
	}
	q := userProfileURL.Query()
	appSecretProof := crypto.HMACSHA256String([]byte(f.Credentials.ClientSecret), []byte(accessTokenResp.AccessToken()))
	q.Set("appsecret_proof", appSecretProof)
	userProfileURL.RawQuery = q.Encode()

	userProfile, err := fetchUserProfile(accessTokenResp, userProfileURL.String())
	if err != nil {
		return
	}
	authInfo.ProviderRawProfile = userProfile

	providerUserInfo, err := f.UserInfoDecoder.DecodeUserInfo(f.ProviderConfig.Type, userProfile)
	if err != nil {
		return
	}
	authInfo.ProviderUserInfo = *providerUserInfo

	return
}

var (
	_ OAuthProvider            = &FacebookImpl{}
	_ NonOpenIDConnectProvider = &FacebookImpl{}
)
