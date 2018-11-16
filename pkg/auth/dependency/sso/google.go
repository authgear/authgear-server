package sso

type GoogleImpl struct {
	Setting Setting
	Config  Config
}

func (f *GoogleImpl) GetAuthURL(params GetURLParams) (url string, err error) {
	params.Options["access_type"] = "offline"
	params.Options["prompt"] = "select_account"
	state := map[string]interface{}{
		"ux_mode":      params.UXMode.String(),
		"callback_url": params.CallbackURL,
		"action":       params.Action,
	}
	if params.UserID != "" {
		state["user_id"] = params.UserID
	}
	return "", nil
}
