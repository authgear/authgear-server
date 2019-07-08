package sso

import (
	"io"
	"net/url"

	"github.com/franela/goreq"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

type AccessTokenResp struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in,omitempty"`
	// Facebook uses "expires" instead of "expires_in"
	RawExpires   int    `json:"expires,omitempty"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
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
