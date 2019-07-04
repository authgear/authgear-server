package sso

import (
	"encoding/json"
	"io"
	"strings"
)

type AuthInfoProcessor interface {
	DecodeAccessTokenResp(r io.Reader) (AccessTokenResp, error)
	ValidateAccessTokenResp(accessTokenResp AccessTokenResp) error
	DecodeUserInfo(p map[string]interface{}) ProviderUserInfo
}

type defaultAuthInfoProcessor struct{}

func newDefaultAuthInfoProcessor() defaultAuthInfoProcessor {
	return defaultAuthInfoProcessor{}
}

func (d defaultAuthInfoProcessor) DecodeAccessTokenResp(r io.Reader) (AccessTokenResp, error) {
	var accessTokenResp AccessTokenResp
	err := json.NewDecoder(r).Decode(&accessTokenResp)
	if err != nil {
		return accessTokenResp, err
	}
	accessTokenResp.Scope = strings.Split(accessTokenResp.RawScope, " ")
	return accessTokenResp, err
}

func (d defaultAuthInfoProcessor) ValidateAccessTokenResp(accessTokenResp AccessTokenResp) error {
	if accessTokenResp.AccessToken == "" {
		err := ssoError{
			code:    MissingAccessToken,
			message: "Missing access token parameter",
		}
		return err
	}

	return nil
}

func (d defaultAuthInfoProcessor) DecodeUserInfo(userProfile map[string]interface{}) ProviderUserInfo {
	id, _ := userProfile["id"].(string)
	email, _ := userProfile["email"].(string)

	return ProviderUserInfo{
		ID:    id,
		Email: email,
	}
}
