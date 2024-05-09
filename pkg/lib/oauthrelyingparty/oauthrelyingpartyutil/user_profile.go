package oauthrelyingpartyutil

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func FetchUserProfile(
	client *http.Client,
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

	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("failed to fetch user profile: unexpected status code: %d", resp.StatusCode)
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&userProfile)
	if err != nil {
		return
	}

	return
}
