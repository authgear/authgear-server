package sso

import (
	"fmt"

	"github.com/franela/goreq"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func fetchUserProfile(
	accessTokenResp AccessTokenResp,
	userProfileURL string,
) (userProfile map[string]interface{}, err error) {
	tokenType := accessTokenResp.TokenType
	accessTokenValue := accessTokenResp.AccessToken
	authorizationHeader := fmt.Sprintf("%s %s", tokenType, accessTokenValue)

	req := goreq.Request{
		Uri:    userProfileURL,
		Method: "GET",
	}
	req.AddHeader("Authorization", authorizationHeader)

	res, err := req.Do()
	if err != nil {
		return
	}

	if res.StatusCode == 401 {
		err = skyerr.NewError(
			skyerr.InvalidCredentials,
			"Invalid access token",
		)
		return
	}

	if res.StatusCode != 200 {
		err = skyerr.NewError(
			skyerr.UnexpectedError,
			"Fail to fetch userinfo",
		)
		return
	}

	err = res.Body.FromJsonTo(&userProfile)
	if err != nil {
		return
	}

	return
}
