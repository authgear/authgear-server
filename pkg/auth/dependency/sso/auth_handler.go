package sso

import (
	"fmt"
)

type authHandler struct {
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

	processAccessToken func(a accessToken) accessToken
	processPrincipalID func(p map[string]interface{}) string
	processAuthData    func(p map[string]interface{}) map[string]interface{}
}

func (h authHandler) handle() (string, error) {
	accessToken, err := fetchAccessToken(
		h.code,
		h.clientID,
		h.urlPrefix,
		h.providerName,
		h.clientSecret,
		h.accessTokenURL,
	)
	if err != nil {
		return "", err
	}

	if h.processAccessToken != nil {
		accessToken = h.processAccessToken(accessToken)
	}

	userProfile, err := fetchUserProfile(accessToken, h.userProfileURL)
	if err != nil {
		return "", err
	}

	if h.processPrincipalID == nil {
		h.processPrincipalID = processPrincipalID
	}
	if h.processAuthData == nil {
		h.processAuthData = processAuthData
	}

	principalID := h.processPrincipalID(userProfile)
	authData := h.processAuthData(userProfile)

	fmt.Printf("principalID = %s\n", principalID)
	fmt.Printf("authData = %v\n", authData)

	return "", nil
}

func processPrincipalID(userProfile map[string]interface{}) string {
	id, ok := userProfile["id"].(string)
	if !ok {
		return ""
	}
	return id
}

func processAuthData(userProfile map[string]interface{}) map[string]interface{} {
	authData := make(map[string]interface{})
	if email, ok := userProfile["email"].(string); ok {
		authData["email"] = email
	}
	return authData
}
