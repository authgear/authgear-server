package wechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	liboauthrelyingparty "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"
)

func init() {
	oauthrelyingparty.RegisterProvider(Type, Wechat{})
}

type AppType string

const (
	AppTypeWeb    AppType = "web"
	AppTypeMobile AppType = "mobile"
)

type ProviderConfig oauthrelyingparty.ProviderConfig

func (c ProviderConfig) AppType() AppType {
	app_type, _ := c["app_type"].(string)
	return AppType(app_type)
}

func (c ProviderConfig) AccountID() string {
	account_id, _ := c["account_id"].(string)
	return account_id
}

func (c ProviderConfig) IsSandboxAccount() bool {
	is_sandbox_account, _ := c["is_sandbox_account"].(bool)
	return is_sandbox_account
}

func (c ProviderConfig) WechatRedirectURIs() []string {
	var out []string
	wechat_redirect_uris, _ := c["wechat_redirect_uris"].([]interface{})
	for _, iface := range wechat_redirect_uris {
		if s, ok := iface.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

const Type = liboauthrelyingparty.TypeWechat

var _ oauthrelyingparty.Provider = Wechat{}
var _ liboauthrelyingparty.BuiltinProvider = Wechat{}

var Schema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"alias": { "type": "string" },
		"type": { "type": "string" },
		"modify_disabled": { "type": "boolean" },
		"client_id": { "type": "string", "minLength": 1 },
		"claims": {
			"type": "object",
			"additionalProperties": false,
			"properties": {
				"email": {
					"type": "object",
					"additionalProperties": false,
					"properties": {
						"assume_verified": { "type": "boolean" },
						"required": { "type": "boolean" }
					}
				}
			}
		},
		"app_type": { "type": "string", "enum": ["mobile", "web"] },
		"account_id": { "type": "string", "format": "wechat_account_id" },
		"is_sandbox_account": { "type": "boolean" },
		"wechat_redirect_uris": { "type": "array", "items": { "type": "string", "format": "uri" } }
	},
	"required": ["alias", "type", "client_id", "app_type", "account_id"]
}
`)

const (
	wechatAuthorizationURL = "https://open.weixin.qq.com/connect/oauth2/authorize"
	// nolint: gosec
	wechatAccessTokenURL = "https://api.weixin.qq.com/sns/oauth2/access_token"
	wechatUserInfoURL    = "https://api.weixin.qq.com/sns/userinfo"
)

type Wechat struct{}

func (Wechat) ValidateProviderConfig(ctx *validation.Context, cfg oauthrelyingparty.ProviderConfig) {
	ctx.AddError(Schema.Validator().ValidateValue(cfg))
}

func (Wechat) SetDefaults(cfg oauthrelyingparty.ProviderConfig) {
	cfg.SetDefaultsModifyDisabledFalse()
	cfg.SetDefaultsEmailClaimConfig(oauthrelyingpartyutil.Email_AssumeVerified_NOT_Required())
}

func (Wechat) ProviderID(cfg oauthrelyingparty.ProviderConfig) oauthrelyingparty.ProviderID {
	// WeChat does NOT support OIDC.
	// In the same Weixin Open Platform account, the user UnionID is unique.
	// The id is scoped to Open Platform account.
	// https://developers.weixin.qq.com/miniprogram/en/dev/framework/open-ability/union-id.html

	wechatCfg := ProviderConfig(cfg)
	account_id := wechatCfg.AccountID()
	is_sandbox_account := wechatCfg.IsSandboxAccount()
	keys := map[string]interface{}{
		"account_id":         account_id,
		"is_sandbox_account": strconv.FormatBool(is_sandbox_account),
	}

	return oauthrelyingparty.NewProviderID(cfg.Type(), keys)
}

func (Wechat) scope() []string {
	// https://developers.weixin.qq.com/doc/offiaccount/OA_Web_Apps/Wechat_webpage_authorization.html
	return []string{"snsapi_userinfo"}
}

func (p Wechat) GetAuthorizationURL(deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetAuthorizationURLOptions) (string, error) {
	return oauthrelyingpartyutil.MakeAuthorizationURL(wechatAuthorizationURL, oauthrelyingpartyutil.AuthorizationURLParams{
		// ClientID is not used by wechat.
		WechatAppID:  deps.ProviderConfig.ClientID(),
		RedirectURI:  param.RedirectURI,
		Scope:        p.scope(),
		ResponseType: oauthrelyingparty.ResponseTypeCode,
		// ResponseMode is unset.
		State: param.State,
		// Prompt is unset.
		// Wechat doesn't support prompt parameter
		// https://developers.weixin.qq.com/doc/oplatform/en/Third-party_Platforms/Official_Accounts/official_account_website_authorization.html
		// Nonce is unset.
	}.Query()), nil
}

func (Wechat) GetUserProfile(deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetUserProfileOptions) (authInfo oauthrelyingparty.UserProfile, err error) {
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

	is_sandbox_account := ProviderConfig(deps.ProviderConfig).IsSandboxAccount()
	var userID string
	if is_sandbox_account {
		if accessTokenResp.UnionID() != "" {
			err = oauthrelyingpartyutil.InvalidConfiguration.New("invalid is_sandbox_account config, WeChat sandbox account should not have union id")
			return
		}
		userID = accessTokenResp.OpenID()
	} else {
		userID = accessTokenResp.UnionID()
	}

	if userID == "" {
		// this may happen if developer misconfigure is_sandbox_account, e.g. sandbox account doesn't have union id
		err = oauthrelyingpartyutil.InvalidConfiguration.New("invalid is_sandbox_account config, missing user id in wechat token response")
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

	body, err := io.ReadAll(resp.Body)
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

	body, err := io.ReadAll(resp.Body)
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
