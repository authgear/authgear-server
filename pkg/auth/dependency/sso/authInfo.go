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
	userID := h.processUserID(userProfile)

	state, err := DecodeState(h.stateJWTSecret, h.encodedState)
	if err != nil {
		return
	}

	authInfo = AuthInfo{
		ProviderName:    h.providerName,
		Action:          state.Action,
		UXMode:          UXModeFromString(state.UXMode),
		UserID:          userID,
		UserProfile:     userProfile,
		AccessTokenResp: accessTokenResp,
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
