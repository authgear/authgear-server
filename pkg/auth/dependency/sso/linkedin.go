package sso

import (
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type LinkedInImpl struct {
	Setting Setting
	Config  Config
}

func (f *LinkedInImpl) GetAuthURL(params GetURLParams) (string, error) {
	if f.Config.ClientID == "" {
		skyErr := skyerr.NewError(skyerr.InvalidArgument, "ClientID is required")
		return "", skyErr
	}
	p := authURLParams{
		providerName:   f.Config.Name,
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

func (f *LinkedInImpl) HandleAuthzResp(code string, scope Scope, encodedState string) (string, error) {
	p := authHandlerParams{
		prividerName:   f.Config.Name,
		clientID:       f.Config.ClientID,
		clientSecret:   f.Config.ClientSecret,
		urlPrefix:      f.Setting.URLPrefix,
		code:           code,
		scope:          scope,
		stateJWTSecret: f.Setting.StateJWTSecret,
		encodedState:   encodedState,
		accessTokenURL: AccessTokenURL(f.Config.Name),
	}
	return authHandler(p)
}
