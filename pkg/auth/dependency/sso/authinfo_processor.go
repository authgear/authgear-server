package sso

import (
	"encoding/json"
	"io"
	"strings"
)

type AuthInfoProcessor interface {
	DecodeAccessTokenResp(r io.Reader) (AccessTokenResp, error)
	ValidateAccessTokenResp(accessTokenResp AccessTokenResp) error
	ProcessUserID(p map[string]interface{}) string
	ProcessAuthData(p map[string]interface{}) map[string]interface{}
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
			message: " Missing access token parameter",
		}
		return err
	}

	return nil
}

func (d defaultAuthInfoProcessor) ProcessUserID(userProfile map[string]interface{}) string {
	id, ok := userProfile["id"].(string)
	if !ok {
		return ""
	}
	return id
}

func (d defaultAuthInfoProcessor) ProcessAuthData(userProfile map[string]interface{}) (authData map[string]interface{}) {
	authData = make(map[string]interface{})
	email, ok := userProfile["email"].(string)
	if ok {
		authData["email"] = email
	}
	return
}
