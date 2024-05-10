package sso

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/wechat"
)

const (
	wechatAuthorizationURL = "https://open.weixin.qq.com/connect/oauth2/authorize"
	// nolint: gosec
	wechatAccessTokenURL = "https://api.weixin.qq.com/sns/oauth2/access_token"
	wechatUserInfoURL    = "https://api.weixin.qq.com/sns/userinfo"
)

type WechatImpl struct{}

func (w *WechatImpl) GetAuthorizationURL(deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetAuthorizationURLOptions) (string, error) {
	return oauthrelyingpartyutil.MakeAuthorizationURL(wechatAuthorizationURL, oauthrelyingpartyutil.AuthorizationURLParams{
		// ClientID is not used by wechat.
		WechatAppID:  deps.ProviderConfig.ClientID(),
		RedirectURI:  param.RedirectURI,
		Scope:        deps.ProviderConfig.Scope(),
		ResponseType: oauthrelyingparty.ResponseTypeCode,
		// ResponseMode is unset.
		State: param.State,
		// Prompt is unset.
		// Wechat doesn't support prompt parameter
		// https://developers.weixin.qq.com/doc/oplatform/en/Third-party_Platforms/Official_Accounts/official_account_website_authorization.html
		// Nonce is unset.
	}.Query()), nil
}

func (w *WechatImpl) GetUserProfile(deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetUserProfileOptions) (authInfo oauthrelyingparty.UserProfile, err error) {
	accessTokenResp, err := wechatFetchAccessTokenResp(
		deps.HTTPClient,
		param.Code,
		deps.ProviderConfig.ClientID(),
		deps.ClientSecret,
	)
	if err != nil {
		return
	}

	rawProfile, err := wechatFetchUserProfile(deps.HTTPClient, accessTokenResp)
	if err != nil {
		return
	}

	is_sandbox_account := wechat.ProviderConfig(deps.ProviderConfig).IsSandboxAccount()
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

	return
}

var (
	_ OAuthProvider = &WechatImpl{}
)

type wechatOAuthErrorResp struct {
	ErrorCode int    `json:"errcode"`
	ErrorMsg  string `json:"errmsg"`
}

func (r *wechatOAuthErrorResp) AsError() error {
	return fmt.Errorf("wechat: %d: %s", r.ErrorCode, r.ErrorMsg)
}

type wechatAccessTokenResp map[string]interface{}

func (r wechatAccessTokenResp) AccessToken() string {
	accessToken, ok := r["access_token"].(string)
	if ok {
		return accessToken
	}
	return ""
}

func (r wechatAccessTokenResp) OpenID() string {
	openid, ok := r["openid"].(string)
	if ok {
		return openid
	}
	return ""
}

func (r wechatAccessTokenResp) UnionID() string {
	unionid, ok := r["unionid"].(string)
	if ok {
		return unionid
	}
	return ""
}

type wechatUserInfoResp map[string]interface{}

func (r wechatUserInfoResp) OpenID() string {
	openid, ok := r["openid"].(string)
	if ok {
		return openid
	}
	return ""
}

func wechatFetchAccessTokenResp(
	client *http.Client,
	code string,
	appid string,
	secret string,
) (r wechatAccessTokenResp, err error) {
	v := url.Values{}
	v.Set("grant_type", "authorization_code")
	v.Add("code", code)
	v.Add("appid", appid)
	v.Add("secret", secret)

	resp, err := client.PostForm(wechatAccessTokenURL, v)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return
	}

	// wechat always return 200
	// to know if there is error, we need to parse the response body
	if resp.StatusCode != 200 {
		err = fmt.Errorf("wechat: unexpected status code: %d", resp.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.NewDecoder(bytes.NewReader(body)).Decode(&r)
	if err != nil {
		return
	}
	if r.AccessToken() == "" {
		// failed to obtain access token, parse the error response
		var errResp wechatOAuthErrorResp
		err = json.NewDecoder(bytes.NewReader(body)).Decode(&errResp)
		if err != nil {
			return
		}
		err = errResp.AsError()
		return
	}
	return
}

func wechatFetchUserProfile(
	client *http.Client,
	accessTokenResp wechatAccessTokenResp,
) (userProfile wechatUserInfoResp, err error) {
	v := url.Values{}
	v.Set("openid", accessTokenResp.OpenID())
	v.Set("access_token", accessTokenResp.AccessToken())
	v.Set("lang", "en")

	resp, err := client.PostForm(wechatUserInfoURL, v)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return
	}

	// wechat always return 200
	// to know if there is error, we need to parse the response body
	if resp.StatusCode != 200 {
		err = fmt.Errorf("wechat: unexpected status code: %d", resp.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.NewDecoder(bytes.NewReader(body)).Decode(&userProfile)
	if err != nil {
		return
	}
	if userProfile.OpenID() == "" {
		// failed to obtain id from user info, parse the error response
		var errResp wechatOAuthErrorResp
		err = json.NewDecoder(bytes.NewReader(body)).Decode(&errResp)
		if err != nil {
			return
		}
		err = errResp.AsError()
		return
	}

	return
}
