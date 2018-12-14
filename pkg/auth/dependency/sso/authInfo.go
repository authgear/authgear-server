package sso

import (
	"encoding/json"
	"io"
	"strings"
)

type authInfoProcessor interface {
	decodeAccessTokenResp(r io.Reader) (AccessTokenResp, error)
	validateAccessTokenResp(accessTokenResp AccessTokenResp) error
	processUserID(p map[string]interface{}) string
	processAuthData(p map[string]interface{}) map[string]interface{}
}

type defaultAuthInfoProcessor struct{}

func newDefaultAuthInfoProcessor() defaultAuthInfoProcessor {
	return defaultAuthInfoProcessor{}
}

type getAuthInfoRequest struct {
	providerName   string
	clientID       string
	clientSecret   string
	urlPrefix      string
	code           string
	scope          Scope
	stateJWTSecret string
	encodedState   string
	accessTokenURL string
	userProfileURL string
	processor      authInfoProcessor
}

func (h getAuthInfoRequest) getAuthInfo() (authInfo AuthInfo, err error) {
	r, err := fetchAccessTokenResp(
		h.code,
		h.clientID,
		h.urlPrefix,
		h.providerName,
		h.clientSecret,
		h.accessTokenURL,
	)
	if err != nil {
		return
	}

	accessTokenResp, err := h.processor.decodeAccessTokenResp(r)
	if err != nil {
		return
	}

	err = h.processor.validateAccessTokenResp(accessTokenResp)
	if err != nil {
		return
	}

	userProfile, err := fetchUserProfile(accessTokenResp, h.userProfileURL)
	if err != nil {
		return
	}

	userID := h.processor.processUserID(userProfile)
	// TODO: process process_userinfo_hook
	authData := h.processor.processAuthData(userProfile)

	state, err := DecodeState(h.stateJWTSecret, h.encodedState)
	if err != nil {
		return
	}

	authInfo = AuthInfo{
		ProviderName:            h.providerName,
		State:                   state,
		ProviderUserID:          userID,
		ProviderUserProfile:     userProfile,
		ProviderAccessTokenResp: accessTokenResp,
		ProviderAuthData:        authData,
	}

	return
}

func (d defaultAuthInfoProcessor) decodeAccessTokenResp(r io.Reader) (AccessTokenResp, error) {
	var accessTokenResp AccessTokenResp
	err := json.NewDecoder(r).Decode(&accessTokenResp)
	if err != nil {
		return accessTokenResp, err
	}
	accessTokenResp.Scope = strings.Split(accessTokenResp.RawScope, " ")
	return accessTokenResp, err
}

func (d defaultAuthInfoProcessor) validateAccessTokenResp(accessTokenResp AccessTokenResp) error {
	if accessTokenResp.AccessToken == "" {
		err := ssoError{
			code:    MissingAccessToken,
			message: " Missing access token parameter",
		}
		return err
	}

	return nil
}

func (d defaultAuthInfoProcessor) processUserID(userProfile map[string]interface{}) string {
	id, ok := userProfile["id"].(string)
	if !ok {
		return ""
	}
	return id
}

func (d defaultAuthInfoProcessor) processAuthData(userProfile map[string]interface{}) (authData map[string]interface{}) {
	authData = make(map[string]interface{})
	email, ok := userProfile["email"].(string)
	if ok {
		authData["email"] = email
	}
	return
}
