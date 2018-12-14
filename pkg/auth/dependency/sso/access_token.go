package sso

import (
	"net/url"

	"github.com/franela/goreq"
)

type AccessTokenResp struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in,omitempty"`
	// Facebook uses "expires" instead of "expires_in"
	RawExpires   int    `json:"expires,omitempty"`
	Scope        Scope  `json:"-"`
	RawScope     string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
}

func fetchAccessTokenResp(
	code string,
	clientID string,
	urlPrefix string,
	providerName string,
	clientSecret string,
	accessTokenURL string,
) (resp []byte, err error) {
	v := url.Values{}
	v.Set("grant_type", "authorization_code")
	v.Add("code", code)
	v.Add("redirect_uri", RedirectURI(urlPrefix, providerName))
	v.Add("client_id", clientID)
	v.Add("client_secret", clientSecret)

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
		var body string
		body, err = res.Body.ToString()
		if err != nil {
			return
		}
		resp = []byte(body)
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
