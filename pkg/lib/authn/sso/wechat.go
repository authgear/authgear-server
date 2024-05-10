package sso

import (
	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/wechat"
)

const (
	wechatAuthorizationURL string = "https://open.weixin.qq.com/connect/oauth2/authorize"
)

type WechatImpl struct {
	ProviderConfig               oauthrelyingparty.ProviderConfig
	ClientSecret                 string
	StandardAttributesNormalizer StandardAttributesNormalizer
	HTTPClient                   OAuthHTTPClient
}

func (w *WechatImpl) Config() oauthrelyingparty.ProviderConfig {
	return w.ProviderConfig
}

func (w *WechatImpl) GetAuthorizationURL(param oauthrelyingparty.GetAuthorizationURLOptions) (string, error) {
	return oauthrelyingpartyutil.MakeAuthorizationURL(wechatAuthorizationURL, oauthrelyingpartyutil.AuthorizationURLParams{
		// ClientID is not used by wechat.
		WechatAppID:  w.ProviderConfig.ClientID(),
		RedirectURI:  param.RedirectURI,
		Scope:        w.ProviderConfig.Scope(),
		ResponseType: oauthrelyingparty.ResponseTypeCode,
		// ResponseMode is unset.
		State: param.State,
		// Prompt is unset.
		// Wechat doesn't support prompt parameter
		// https://developers.weixin.qq.com/doc/oplatform/en/Third-party_Platforms/Official_Accounts/official_account_website_authorization.html
		// Nonce is unset.
	}.Query()), nil
}

func (w *WechatImpl) GetAuthInfo(param GetAuthInfoParam) (authInfo AuthInfo, err error) {
	accessTokenResp, err := wechatFetchAccessTokenResp(
		w.HTTPClient,
		param.Code,
		w.ProviderConfig.ClientID(),
		w.ClientSecret,
	)
	if err != nil {
		return
	}

	rawProfile, err := wechatFetchUserProfile(w.HTTPClient, accessTokenResp)
	if err != nil {
		return
	}

	is_sandbox_account := wechat.ProviderConfig(w.ProviderConfig).IsSandboxAccount()
	var userID string
	if is_sandbox_account {
		if accessTokenResp.UnionID() != "" {
			err = InvalidConfiguration.New("invalid is_sandbox_account config, WeChat sandbox account should not have union id")
			return
		}
		userID = accessTokenResp.OpenID()
	} else {
		userID = accessTokenResp.UnionID()
	}

	if userID == "" {
		// this may happen if developer misconfigure is_sandbox_account, e.g. sandbox account doesn't have union id
		err = InvalidConfiguration.New("invalid is_sandbox_account config, missing user id in wechat token response")
		return
	}

	// https://developers.weixin.qq.com/doc/offiaccount/User_Management/Get_users_basic_information_UnionID.html
	// Here is an example of how the raw profile looks like.
	// {
	//     "sex": 0,
	//     "city": "",
	//     "openid": "redacted",
	//     "country": "",
	//     "language": "zh_CN",
	//     "nickname": "John Doe",
	//     "province": "",
	//     "privilege": [],
	//     "headimgurl": ""
	// }
	var gender string
	if sex, ok := rawProfile["sex"].(float64); ok {
		if sex == 1 {
			gender = "male"
		} else if sex == 2 {
			gender = "female"
		}
	}

	name, _ := rawProfile["nickname"].(string)
	locale, _ := rawProfile["language"].(string)

	authInfo.ProviderRawProfile = rawProfile
	authInfo.ProviderUserID = userID

	// Claims.Email.Required is not respected because wechat does not return the email claim.
	authInfo.StandardAttributes = stdattrs.T{
		stdattrs.Name:   name,
		stdattrs.Locale: locale,
		stdattrs.Gender: gender,
	}.WithNameCopiedToGivenName()

	err = w.StandardAttributesNormalizer.Normalize(authInfo.StandardAttributes)
	if err != nil {
		return
	}

	return
}

var (
	_ OAuthProvider = &WechatImpl{}
)
