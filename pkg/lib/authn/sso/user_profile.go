package sso

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
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
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		err = errorutil.WithSecondaryError(
			NewSSOFailed(NetworkFailed, "failed to connect authorization server"),
			err,
		)
		return
	}

	if resp.StatusCode == 401 {
		err = NewSSOFailed(SSOUnauthorized, "oauth failed")
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&userProfile)
	if err != nil {
		return
	}

	return
}
