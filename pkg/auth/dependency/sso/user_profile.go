package sso

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func fetchUserProfile(
	accessTokenResp AccessTokenResp,
	userProfileURL string,
) (userProfile map[string]interface{}, err error) {
	tokenType := accessTokenResp.TokenType()
	accessTokenValue := accessTokenResp.AccessToken()
	authorizationHeader := fmt.Sprintf("%s %s", tokenType, accessTokenValue)

	req, err := http.NewRequest(http.MethodGet, userProfileURL, nil)
	if err != nil {
		return
	}
	req.Header.Add("Authorization", authorizationHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		err = skyerr.NewError(
			skyerr.InvalidCredentials,
			"Invalid access token",
		)
		return
	}

	if resp.StatusCode != 200 {
		err = skyerr.NewError(
			skyerr.UnexpectedError,
			"Fail to fetch userinfo",
		)
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&userProfile)
	if err != nil {
		return
	}

	return
}
