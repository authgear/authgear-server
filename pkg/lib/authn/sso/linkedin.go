package sso

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

const (
	linkedinAuthorizationURL string = "https://www.linkedin.com/oauth/v2/authorization"
	// nolint: gosec
	linkedinTokenURL   string = "https://www.linkedin.com/oauth/v2/accessToken"
	linkedinMeURL      string = "https://api.linkedin.com/v2/me"
	linkedinContactURL string = "https://api.linkedin.com/v2/clientAwareMemberHandles?q=members&projection=(elements*(primary,type,handle~))"
)

type LinkedInImpl struct {
	RedirectURL              RedirectURLProvider
	ProviderConfig           config.OAuthSSOProviderConfig
	Credentials              config.OAuthClientCredentialsItem
	LoginIDNormalizerFactory LoginIDNormalizerFactory
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
		prompt:      f.GetPrompt(param.Prompt),
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

	providerUserInfo := decodeLinkedIn(combinedResponse)
	if providerUserInfo.Email != "" {
		var email string
		normalizer := f.LoginIDNormalizerFactory.NormalizerWithLoginIDType(config.LoginIDKeyTypeEmail)
		email, err = normalizer.Normalize(providerUserInfo.Email)
		if err != nil {
			return
		}
		providerUserInfo.Email = email
	}

	authInfo.ProviderRawProfile = combinedResponse
	authInfo.ProviderUserInfo = providerUserInfo
	return
}

func (f *LinkedInImpl) GetPrompt(prompt []string) []string {
	// linkedin doesn't support prompt parameter
	// ref: https://docs.microsoft.com/en-us/linkedin/shared/authentication/authorization-code-flow?tabs=HTTPS#step-2-request-an-authorization-code
	return []string{}
}

func decodeLinkedIn(userInfo map[string]interface{}) ProviderUserInfo {
	profile := userInfo["profile"].(map[string]interface{})
	id := profile["id"].(string)

	email := ""
	primaryContact := userInfo["primary_contact"].(map[string]interface{})
	elements := primaryContact["elements"].([]interface{})
	for _, e := range elements {
		element := e.(map[string]interface{})
		if primary, ok := element["primary"].(bool); !ok || !primary {
			continue
		}
		if typ, ok := element["type"].(string); !ok || typ != "EMAIL" {
			continue
		}
		handleTilde, ok := element["handle~"].(map[string]interface{})
		if !ok {
			continue
		}
		email, _ = handleTilde["emailAddress"].(string)
	}

	return ProviderUserInfo{
		ID:    id,
		Email: email,
	}
}

var (
	_ OAuthProvider            = &LinkedInImpl{}
	_ NonOpenIDConnectProvider = &LinkedInImpl{}
)
