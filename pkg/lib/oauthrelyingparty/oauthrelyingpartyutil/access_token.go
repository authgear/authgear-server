package oauthrelyingpartyutil

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
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
		// LinkedIn does not include token_type in the response.
		return "Bearer"
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

func FetchAccessTokenResp(
	client *http.Client,
	code string,
	accessTokenURL string,
	redirectURL string,
	clientID string,
	clientSecret string,
) (r AccessTokenResp, err error) {
	v := url.Values{}
	v.Set("grant_type", "authorization_code")
	v.Add("code", code)
	v.Add("redirect_uri", redirectURL)
	v.Add("client_id", clientID)
	v.Add("client_secret", clientSecret)

	// nolint: gosec
	resp, err := client.PostForm(accessTokenURL, v)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return
	}

	if resp.StatusCode == 200 {
		err = json.NewDecoder(resp.Body).Decode(&r)
		if err != nil {
			return
		}
	} else { // normally 400 Bad Request
		var errResp oauthrelyingparty.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&errResp)
		if err != nil {
			return
		}
		err = ErrorResponseAsError(errResp)
	}

	return
}
