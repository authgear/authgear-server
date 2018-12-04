package sso

import (
	"net/url"
	"strings"

	"github.com/franela/goreq"
)

type authHandlerParams struct {
	prividerName   string
	clientID       string
	clientSecret   string
	urlPrefix      string
	code           string
	scope          Scope
	stateJWTSecret string
	encodedState   string
	accessTokenURL string
}

type accessTokenResp struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in,omitempty"`
	// Facebook uses "expires" instead of "expires_in"
	Expires      int    `json:"expires,omitempty"`
	Scope        Scope  `json:"scope"`
	RefreshToken string `json:"refresh_token"`
}

func fetchToken(params authHandlerParams) (accessTokenResp accessTokenResp, err error) {
	v := url.Values{}
	v.Set("grant_type", "authorization_code")
	v.Add("code", params.code)
	v.Add("redirect_uri", RedirectURI(params.urlPrefix, params.prividerName))
	v.Add("client_id", params.clientID)
	v.Add("client_secret", params.clientSecret)

	res, err := goreq.Request{
		Uri:         params.accessTokenURL,
		Method:      "POST",
		Body:        v.Encode(),
		ContentType: "application/x-www-form-urlencoded; charset=UTF-8",
	}.Do()

	if err != nil {
		return
	}

	if res.StatusCode == 200 {
		err = res.Body.FromJsonTo(&accessTokenResp)
		if err != nil {
			return
		}
	} else { // normally 400 Bad Request
		var errResp ErrorResp
		err = res.Body.FromJsonTo(&errResp)
		if err != nil {
			return
		}
		err = RespToError(errResp)
	}

	return
}

func authHandler(params authHandlerParams) (string, error) {
	accessTokenResp, err := fetchToken(params)
	if err != nil {
		return "", err
	}

	if params.prividerName == "facebook" {
		// need some special handlings for facebook sso login
		if accessTokenResp.ExpiresIn == 0 && accessTokenResp.Expires != 0 {
			accessTokenResp.ExpiresIn = accessTokenResp.Expires
		}
		if strings.ToLower(accessTokenResp.TokenType) == "bearer" {
			accessTokenResp.TokenType = "Bearer"
		}
	}

	return "", nil
}
