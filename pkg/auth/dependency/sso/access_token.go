package sso

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

type AccessTokenResp map[string]interface{}

func NewBearerAccessTokenResp(accessToken string) AccessTokenResp {
	return AccessTokenResp{
		"token_type":   "Bearer",
		"access_token": accessToken,
	}
}

func (r AccessTokenResp) IDToken() string {
	idToken, ok := r["id_token"].(string)
	if ok {
		return idToken
	}
	return ""
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

func (r AccessTokenResp) Validate() error {
	if r.AccessToken() == "" {
		err := ssoError{
			code:    MissingAccessToken,
			message: "Missing access token parameter",
		}
		return err
	}

	return nil
}

func fetchAccessTokenResp(
	code string,
	accessTokenURL string,
	urlPrefix *url.URL,
	oauthConfig config.OAuthConfiguration,
	providerConfig config.OAuthProviderConfiguration,
) (r AccessTokenResp, err error) {
	v := url.Values{}
	v.Set("grant_type", "authorization_code")
	v.Add("code", code)
	v.Add("redirect_uri", redirectURI(urlPrefix, providerConfig))
	v.Add("client_id", providerConfig.ClientID)
	v.Add("client_secret", providerConfig.ClientSecret)

	// nolint: gosec
	resp, err := http.PostForm(accessTokenURL, v)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		err = json.NewDecoder(resp.Body).Decode(&r)
		if err != nil {
			return
		}
	} else { // normally 400 Bad Request
		var errResp ErrorResp
		err = json.NewDecoder(resp.Body).Decode(&errResp)
		if err != nil {
			return
		}
		err = respToError(errResp)
	}

	return
}
