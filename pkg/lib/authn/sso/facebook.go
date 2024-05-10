package sso

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
	"github.com/authgear/authgear-server/pkg/util/crypto"
)

const (
	facebookAuthorizationURL string = "https://www.facebook.com/v11.0/dialog/oauth"
	// nolint: gosec
	facebookTokenURL    string = "https://graph.facebook.com/v11.0/oauth/access_token"
	facebookUserInfoURL string = "https://graph.facebook.com/v11.0/me?fields=id,email,first_name,last_name,middle_name,name,name_format,picture,short_name"
)

type FacebookImpl struct{}

func (f *FacebookImpl) GetAuthorizationURL(deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetAuthorizationURLOptions) (string, error) {
	return oauthrelyingpartyutil.MakeAuthorizationURL(facebookAuthorizationURL, oauthrelyingpartyutil.AuthorizationURLParams{
		ClientID:     deps.ProviderConfig.ClientID(),
		RedirectURI:  param.RedirectURI,
		Scope:        deps.ProviderConfig.Scope(),
		ResponseType: oauthrelyingparty.ResponseTypeCode,
		// ResponseMode is unset
		State: param.State,
		// Prompt is unset.
		// Facebook doesn't support prompt parameter
		// https://developers.facebook.com/docs/facebook-login/manually-build-a-login-flow/

		// Nonce is unset
	}.Query()), nil
}

func (f *FacebookImpl) GetUserProfile(deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetUserProfileOptions) (authInfo oauthrelyingparty.UserProfile, err error) {
	authInfo = oauthrelyingparty.UserProfile{}

	accessTokenResp, err := oauthrelyingpartyutil.FetchAccessTokenResp(
		deps.HTTPClient,
		param.Code,
		facebookTokenURL,
		param.RedirectURI,
		deps.ProviderConfig.ClientID(),
		deps.ClientSecret,
	)
	if err != nil {
		return
	}

	userProfileURL, err := url.Parse(facebookUserInfoURL)
	if err != nil {
		return
	}
	q := userProfileURL.Query()
	appSecretProof := crypto.HMACSHA256String([]byte(deps.ClientSecret), []byte(accessTokenResp.AccessToken()))
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

	userProfile, err := oauthrelyingpartyutil.FetchUserProfile(deps.HTTPClient, accessTokenResp, userProfileURL.String())
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

	authInfo.ProviderUserID = id
	emailRequired := deps.ProviderConfig.EmailClaimConfig().Required()
	stdAttrs, err := stdattrs.Extract(map[string]interface{}{
		stdattrs.Email:      email,
		stdattrs.GivenName:  firstName,
		stdattrs.FamilyName: lastName,
		stdattrs.Name:       name,
		stdattrs.Nickname:   shortName,
		stdattrs.Picture:    picture,
	}, stdattrs.ExtractOptions{
		EmailRequired: emailRequired,
	})
	if err != nil {
		return
	}
	authInfo.StandardAttributes = stdAttrs

	return
}

var (
	_ OAuthProvider = &FacebookImpl{}
)
