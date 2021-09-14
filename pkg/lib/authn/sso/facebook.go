package sso

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/crypto"
)

const (
	facebookAuthorizationURL string = "https://www.facebook.com/v11.0/dialog/oauth"
	// nolint: gosec
	facebookTokenURL    string = "https://graph.facebook.com/v11.0/oauth/access_token"
	facebookUserInfoURL string = "https://graph.facebook.com/v11.0/me?fields=id,email,first_name,last_name,middle_name,name,name_format,picture,short_name"
)

type FacebookImpl struct {
	RedirectURL                  RedirectURLProvider
	ProviderConfig               config.OAuthSSOProviderConfig
	Credentials                  config.OAuthClientCredentialsItem
	StandardAttributesNormalizer StandardAttributesNormalizer
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
		prompt:      f.GetPrompt(param.Prompt),
	}
	return authURL(p)
}

func (f *FacebookImpl) GetAuthInfo(r OAuthAuthorizationResponse, param GetAuthInfoParam) (authInfo AuthInfo, err error) {
	return f.NonOpenIDConnectGetAuthInfo(r, param)
}

func (f *FacebookImpl) NonOpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, _ GetAuthInfoParam) (authInfo AuthInfo, err error) {
	authInfo = AuthInfo{}

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

	userProfileURL, err := url.Parse(facebookUserInfoURL)
	if err != nil {
		return
	}
	q := userProfileURL.Query()
	appSecretProof := crypto.HMACSHA256String([]byte(f.Credentials.ClientSecret), []byte(accessTokenResp.AccessToken()))
	q.Set("appsecret_proof", appSecretProof)
	userProfileURL.RawQuery = q.Encode()

	// Here is the refacted user profile of Louis' facebook account.
	// {
	//   "id": "redacted",
	//   "email": "redacted",
	//   "first_name": "Jonathan",
	//   "last_name": "Doe",
	//   "name": "Johnathan Doe",
	//   "name_format": "{first} {last}",
	//   "picture": {
	//     "data": {
	//       "height": 50,
	//       "is_silhouette": true,
	//       "url": "http://example.com",
	//       "width": 50
	//     }
	//   },
	//   "short_name": "John"
	// }

	userProfile, err := fetchUserProfile(accessTokenResp, userProfileURL.String())
	if err != nil {
		return
	}
	authInfo.ProviderRawProfile = userProfile

	id, _ := userProfile["id"].(string)
	email, _ := userProfile["email"].(string)
	firstName, _ := userProfile["first_name"].(string)
	lastName, _ := userProfile["last_name"].(string)
	name, _ := userProfile["name"].(string)
	shortName, _ := userProfile["short_name"].(string)
	var picture string
	if pictureObj, ok := userProfile["picture"].(map[string]interface{}); ok {
		if data, ok := pictureObj["data"].(map[string]interface{}); ok {
			if url, ok := data["url"].(string); ok {
				picture = url
			}
		}
	}

	authInfo.StandardAttributes = stdattrs.T{
		stdattrs.Sub:        id,
		stdattrs.Email:      email,
		stdattrs.GivenName:  firstName,
		stdattrs.FamilyName: lastName,
		stdattrs.Name:       name,
		stdattrs.Nickname:   shortName,
		stdattrs.Picture:    picture,
	}

	err = f.StandardAttributesNormalizer.Normalize(authInfo.StandardAttributes)
	if err != nil {
		return
	}

	return
}

func (f *FacebookImpl) GetPrompt(prompt []string) []string {
	// facebook doesn't support prompt parameter
	// https://developers.facebook.com/docs/facebook-login/manually-build-a-login-flow/
	return []string{}
}

var (
	_ OAuthProvider            = &FacebookImpl{}
	_ NonOpenIDConnectProvider = &FacebookImpl{}
)
