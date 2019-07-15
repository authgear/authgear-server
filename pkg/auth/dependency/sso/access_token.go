package sso

import (
	"io"
	"net/url"
	"strconv"
	"strings"

	"github.com/franela/goreq"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

type AccessTokenResp map[string]interface{}

func NewBearerAccessTokenResp(accessToken string) AccessTokenResp {
	return AccessTokenResp{
		"token_type":   "Bearer",
		"access_token": accessToken,
	}
}

func (r AccessTokenResp) AccessToken() string {
	accessToken, ok := r["access_token"].(string)
	if ok {
		return accessToken
	}
	return ""
}

func (r AccessTokenResp) ExpiresIn() int {
	expires, hasExpires := r["expires"]
	expiresIn, hasExpiresIn := r["expires_in"]

	// Facebook use "expires" instead of "expires_in"
	if hasExpires && !hasExpiresIn {
		expiresIn = expires
	}

	switch v := expiresIn.(type) {
	// Azure AD v2 uses string instead of number
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			return 0
		}
		return i
	case float64:
		return int(v)
	default:
		return 0
	}
}

func (r AccessTokenResp) TokenType() string {
	tokenType, ok := r["token_type"].(string)
	if !ok {
		return ""
	}
	tokenType = strings.ToLower(tokenType)
	switch tokenType {
	case "basic":
		return "Basic"
	case "digest":
		return "Digest"
	case "bearer":
		return "Bearer"
	// We do not care about other less common schemes.
	default:
		return tokenType
	}
}

func fetchAccessTokenResp(
	code string,
	accessTokenURL string,
	oauthConfig config.OAuthConfiguration,
	providerConfig config.OAuthProviderConfiguration,
) (r io.Reader, err error) {
	v := url.Values{}
	v.Set("grant_type", "authorization_code")
	v.Add("code", code)
	v.Add("redirect_uri", RedirectURI(oauthConfig, providerConfig))
	v.Add("client_id", providerConfig.ClientID)
	v.Add("client_secret", providerConfig.ClientSecret)

	res, err := goreq.Request{
		Uri:         accessTokenURL,
		Method:      "POST",
		Body:        v.Encode(),
		ContentType: "application/x-www-form-urlencoded; charset=UTF-8",
	}.Do()

	if err != nil {
		return
	}

	if res.StatusCode == 200 {
		r = res.Body
	} else { // normally 400 Bad Request
		var errResp ErrorResp
		err = res.Body.FromJsonTo(&errResp)
		if err != nil {
			return
		}
		err = respToError(errResp)
	}

	return
}
