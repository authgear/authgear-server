package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	liboauthrelyingparty "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/validation"
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

const (
	// As of 2025-06-23, WeChat offers 2 ways to integrate.
	// 1. Use the WeChat hosted QR page.
	//    Basically you just redirect the user to that page.
	//    And the user is supposed to scan the QR code there.
	// 2. Load the WeChat JS library to in our HTML, and use the library to draw a QR code.
	//    This needs more work so it is not used.
	wechatQRCodePageURL = "https://open.weixin.qq.com/connect/qrconnect"
	// nolint: gosec
	wechatAccessTokenURL = "https://api.weixin.qq.com/sns/oauth2/access_token"
	wechatUserInfoURL    = "https://api.weixin.qq.com/sns/userinfo"
)

type Wechat struct{}

func (Wechat) GetJSONSchema() map[string]interface{} {
	builder := validation.SchemaBuilder{}
	builder.Type(validation.TypeObject)
	builder.Properties().
		Property("type", validation.SchemaBuilder{}.Type(validation.TypeString)).
		Property("client_id", validation.SchemaBuilder{}.Type(validation.TypeString).MinLength(1)).
		Property("claims", validation.SchemaBuilder{}.Type(validation.TypeObject).
			AdditionalPropertiesFalse().
			Properties().
			Property("email", validation.SchemaBuilder{}.Type(validation.TypeObject).
				AdditionalPropertiesFalse().Properties().
				Property("assume_verified", validation.SchemaBuilder{}.Type(validation.TypeBoolean)).
				Property("required", validation.SchemaBuilder{}.Type(validation.TypeBoolean)),
			),
		).
		Property("app_type", validation.SchemaBuilder{}.Type(validation.TypeString).Enum("mobile", "web")).
		Property("account_id", validation.SchemaBuilder{}.Type(validation.TypeString).Format("wechat_account_id")).
		Property("is_sandbox_account", validation.SchemaBuilder{}.Type(validation.TypeBoolean)).
		Property("wechat_redirect_uris", validation.SchemaBuilder{}.Type(validation.TypeArray).
			Items(validation.SchemaBuilder{}.Type(validation.TypeString).Format("uri")),
		)
	builder.Required("type", "client_id", "app_type", "account_id")
	return builder
}

func (Wechat) SetDefaults(cfg oauthrelyingparty.ProviderConfig) {
	cfg.SetDefaultsEmailClaimConfig(oauthrelyingpartyutil.Email_AssumeVerified_NOT_Required())
}

func (Wechat) ProviderID(cfg oauthrelyingparty.ProviderConfig) oauthrelyingparty.ProviderID {
	// WeChat does NOT support OIDC.
	// In the same Weixin Open Platform account, the user UnionID is unique.
	// The id is scoped to Open Platform account.
	// https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Authorized_Interface_Calling_UnionID.html

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
	// According to the documentation as of 2025-06-23, the only value of scope is `scope=snsapi_login`.
	// https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Wechat_Login.html
	return []string{"snsapi_login"}
}

func (p Wechat) GetAuthorizationURL(ctx context.Context, deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetAuthorizationURLOptions) (string, error) {
	withoutFragment := oauthrelyingpartyutil.MakeAuthorizationURL(wechatQRCodePageURL, oauthrelyingpartyutil.AuthorizationURLParams{
		// The supported query parameters are documented at
		// https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Wechat_Login.html
		// As of 2025-06-23, the supported parameters are:
		// - appid: The appid
		// - redirect_uri: Any URL of the registered domain is allowed.
		// - response_type: It must be `code`.
		// - scope: It must be `snsapi_login`.
		// - state: OAuth 2.0 state parameter.
		// - lang: Either `cn` or en`. If not specified, `cn` is assumed.

		ExtraQuery: url.Values{
			"appid": []string{deps.ProviderConfig.ClientID()},
		},
		RedirectURI:  param.RedirectURI,
		Scope:        p.scope(),
		ResponseType: oauthrelyingparty.ResponseTypeCode,
		State:        param.State,
	}.Query())

	// The doc says the fragment is important.
	// https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Wechat_Login.html
	withFragment := withoutFragment + "#wechat_redirect"
	return withFragment, nil
}

func (Wechat) GetUserProfile(ctx context.Context, deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetUserProfileOptions) (authInfo oauthrelyingparty.UserProfile, err error) {
	code, err := oauthrelyingpartyutil.GetCode(param.Query)
	if err != nil {
		return
	}

	accessTokenResp, err := wechatFetchAccessTokenResp(
		ctx,
		deps.HTTPClient,
		code,
		deps.ProviderConfig.ClientID(),
		deps.ClientSecret,
	)
	if err != nil {
		return
	}

	rawProfile, err := wechatFetchUserProfile(ctx, deps.HTTPClient, accessTokenResp)
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

	// https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Authorized_Interface_Calling_UnionID.html
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
	return &oauthrelyingparty.ErrorResponse{
		Error_:           fmt.Sprintf("%v", r.ErrorCode),
		ErrorDescription: r.ErrorMsg,
	}
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
	ctx context.Context,
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

	resp, err := httputil.PostFormWithContext(ctx, client, wechatAccessTokenURL, v)
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
	ctx context.Context,
	client *http.Client,
	accessTokenResp wechatAccessTokenResp,
) (userProfile wechatUserInfoResp, err error) {
	v := url.Values{}
	v.Set("openid", accessTokenResp.OpenID())
	v.Set("access_token", accessTokenResp.AccessToken())
	v.Set("lang", "en")

	resp, err := httputil.PostFormWithContext(ctx, client, wechatUserInfoURL, v)
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
