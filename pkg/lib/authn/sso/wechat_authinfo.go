package sso

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	// nolint: gosec
	wechatAccessTokenURL = "https://api.weixin.qq.com/sns/oauth2/access_token"
	wechatUserInfoURL    = "https://api.weixin.qq.com/sns/userinfo"
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
