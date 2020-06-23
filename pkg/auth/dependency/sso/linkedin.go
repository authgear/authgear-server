package sso

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/auth/config"
)

const (
	linkedinAuthorizationURL string = "https://www.linkedin.com/oauth/v2/authorization"
	// nolint: gosec
	linkedinTokenURL   string = "https://www.linkedin.com/oauth/v2/accessToken"
	linkedinMeURL      string = "https://api.linkedin.com/v2/me"
	linkedinContactURL string = "https://api.linkedin.com/v2/clientAwareMemberHandles?q=members&projection=(elements*(primary,type,handle~))"
)

type LinkedInImpl struct {
	URLPrefix       *url.URL
	RedirectURLFunc RedirectURLFunc
	ProviderConfig  config.OAuthSSOProviderConfig
	Credentials     config.OAuthClientCredentialsItem
	UserInfoDecoder UserInfoDecoder
}

func (f *LinkedInImpl) Type() config.OAuthSSOProviderType {
	return config.OAuthSSOProviderTypeLinkedIn
}

func (f *LinkedInImpl) GetAuthURL(state State, encodedState string) (string, error) {
	p := authURLParams{
		redirectURI:  f.RedirectURLFunc(f.URLPrefix, f.ProviderConfig),
		clientID:     f.ProviderConfig.ClientID,
		encodedState: encodedState,
		baseURL:      linkedinAuthorizationURL,
	}
	return authURL(p)
}

func (f *LinkedInImpl) GetAuthInfo(r OAuthAuthorizationResponse, state State) (authInfo AuthInfo, err error) {
	return f.NonOpenIDConnectGetAuthInfo(r, state)
}

func (f *LinkedInImpl) NonOpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, state State) (authInfo AuthInfo, err error) {
	accessTokenResp, err := fetchAccessTokenResp(
		r.Code,
		linkedinTokenURL,
		f.RedirectURLFunc(f.URLPrefix, f.ProviderConfig),
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
