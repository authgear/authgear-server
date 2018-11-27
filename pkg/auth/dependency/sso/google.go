package sso

import "github.com/skygeario/skygear-server/pkg/server/skyerr"

type GoogleImpl struct {
	Setting Setting
	Config  Config
}

func (f *GoogleImpl) GetAuthURL(params GetURLParams) (string, error) {
	if f.Config.ClientID == "" {
		skyErr := skyerr.NewError(skyerr.InvalidArgument, "ClientID is required")
		return "", skyErr
	}
	params.Options["access_type"] = "offline"
	params.Options["prompt"] = "select_account"
	p := authURLParams{
		prividerName:   f.Config.Name,
		clientID:       f.Config.ClientID,
		urlPrefix:      f.Setting.URLPrefix,
		scope:          GetScope(params.Scope, f.Config.Scope),
		options:        params.Options,
		stateJWTSecret: f.Setting.StateJWTSecret,
		state:          NewState(params),
		baseURL:        BaseURL(f.Config.Name),
	}
	return authURL(p)
}
