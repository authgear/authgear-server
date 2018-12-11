package sso

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

	processAccessTokenResp func(a AccessTokenResp) AccessTokenResp
	processUserID          func(p map[string]interface{}) string
	processAuthData        func(p map[string]interface{}) map[string]interface{}
}

func (h authHandler) getAuthInfo() (authInfo AuthInfo, err error) {
	accessTokenResp, err := fetchAccessTokenResp(
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

	if h.processAccessTokenResp != nil {
		accessTokenResp = h.processAccessTokenResp(accessTokenResp)
	}

	userProfile, err := fetchUserProfile(accessTokenResp, h.userProfileURL)
	if err != nil {
		return
	}

	if h.processUserID == nil {
		h.processUserID = processUserID
	}
	if h.processAuthData == nil {
		h.processAuthData = processAuthData
	}

	userID := h.processUserID(userProfile)
	// TODO: process process_userinfo_hook
	authData := h.processAuthData(userProfile)

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

func processUserID(userProfile map[string]interface{}) string {
	id, ok := userProfile["id"].(string)
	if !ok {
		return ""
	}
	return id
}

func processAuthData(userProfile map[string]interface{}) (authData map[string]interface{}) {
	email, ok := userProfile["email"].(string)
	if ok {
		authData["email"] = email
	}
	return
}
