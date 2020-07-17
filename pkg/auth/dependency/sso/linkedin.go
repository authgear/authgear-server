package sso

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
)

const (
	linkedinAuthorizationURL string = "https://www.linkedin.com/oauth/v2/authorization"
	// nolint: gosec
	linkedinTokenURL   string = "https://www.linkedin.com/oauth/v2/accessToken"
	linkedinMeURL      string = "https://api.linkedin.com/v2/me"
	linkedinContactURL string = "https://api.linkedin.com/v2/clientAwareMemberHandles?q=members&projection=(elements*(primary,type,handle~))"
)

type LinkedInImpl struct {
	RedirectURL     RedirectURLProvider
	ProviderConfig  config.OAuthSSOProviderConfig
	Credentials     config.OAuthClientCredentialsItem
	UserInfoDecoder UserInfoDecoder
}

func (*LinkedInImpl) Type() config.OAuthSSOProviderType {
	return config.OAuthSSOProviderTypeLinkedIn
}

func (f *LinkedInImpl) Config() config.OAuthSSOProviderConfig {
	return f.ProviderConfig
}

func (f *LinkedInImpl) GetAuthURL(param GetAuthURLParam) (string, error) {
	p := authURLParams{
		redirectURI: f.RedirectURL.SSOCallbackURL(f.ProviderConfig).String(),
		clientID:    f.ProviderConfig.ClientID,
		scope:       f.ProviderConfig.Type.Scope(),
		state:       param.State,
		baseURL:     linkedinAuthorizationURL,
	}
	return authURL(p)
}

func (f *LinkedInImpl) GetAuthInfo(r OAuthAuthorizationResponse, param GetAuthInfoParam) (authInfo AuthInfo, err error) {
	return f.NonOpenIDConnectGetAuthInfo(r, param)
}

func (f *LinkedInImpl) NonOpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, _ GetAuthInfoParam) (authInfo AuthInfo, err error) {
	accessTokenResp, err := fetchAccessTokenResp(
		r.Code,
		linkedinTokenURL,
		f.RedirectURL.SSOCallbackURL(f.ProviderConfig).String(),
		f.ProviderConfig.ClientID,
		f.Credentials.ClientSecret,
	)
	if err != nil {
		return
	}

	meResponse, err := fetchUserProfile(accessTokenResp, linkedinMeURL)
	if err != nil {
		return
	}

	contactResponse, err := fetchUserProfile(accessTokenResp, linkedinContactURL)
	if err != nil {
		return
	}

	combinedResponse := map[string]interface{}{
		"profile":         meResponse,
		"primary_contact": contactResponse,
	}

	providerUserInfo, err := f.UserInfoDecoder.DecodeUserInfo(f.ProviderConfig.Type, combinedResponse)
	if err != nil {
		return
	}

	authInfo.ProviderConfig = f.ProviderConfig
	authInfo.ProviderAccessTokenResp = accessTokenResp
	authInfo.ProviderRawProfile = combinedResponse
	authInfo.ProviderUserInfo = *providerUserInfo
	return
}

var (
	_ OAuthProvider            = &LinkedInImpl{}
	_ NonOpenIDConnectProvider = &LinkedInImpl{}
)
