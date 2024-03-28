package sso

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type MockImpl struct {
	ProviderConfig               config.OAuthSSOProviderConfig
	Credentials                  config.OAuthSSOProviderCredentialsItem
	StandardAttributesNormalizer StandardAttributesNormalizer
}

func (*MockImpl) Type() config.OAuthSSOProviderType {
	return config.OAuthSSOProviderTypeMock
}

func (w *MockImpl) Config() config.OAuthSSOProviderConfig {
	return w.ProviderConfig
}

func (w *MockImpl) GetAuthURL(param GetAuthURLParam) (string, error) {
	// Directly return the redirect URI with code and state
	url, err := url.Parse(param.RedirectURI)
	if err != nil {
		return "", err
	}

	url.Query().Add("code", "mock-code")
	url.Query().Add("state", param.State)

	return url.String(), nil
}

func (w *MockImpl) GetAuthInfo(r OAuthAuthorizationResponse, param GetAuthInfoParam) (AuthInfo, error) {
	return w.NonOpenIDConnectGetAuthInfo(r, param)
}

func (w *MockImpl) NonOpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, _ GetAuthInfoParam) (authInfo AuthInfo, err error) {
	authInfo.ProviderRawProfile = make(map[string]interface{})
	authInfo.ProviderUserID = "mock-user-id"

	authInfo.StandardAttributes = stdattrs.T{
		stdattrs.Name:   "Mock User",
		stdattrs.Locale: "en",
		stdattrs.Gender: "unknown",
	}.WithNameCopiedToGivenName()

	err = w.StandardAttributesNormalizer.Normalize(authInfo.StandardAttributes)
	if err != nil {
		return
	}

	return
}

func (w *MockImpl) GetPrompt(prompt []string) []string {
	// mock doesn't support prompt parameter
	// ref: https://developers.weixin.qq.com/doc/oplatform/en/Third-party_Platforms/Official_Accounts/official_account_website_authorization.html
	return []string{}
}

var (
	_ OAuthProvider            = &MockImpl{}
	_ NonOpenIDConnectProvider = &MockImpl{}
)
