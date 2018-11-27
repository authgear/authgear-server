package sso

import (
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type MockSSOProverImpl struct {
	BaseURL string
	Setting Setting
	Config  Config
}

func (f *MockSSOProverImpl) GetAuthURL(params GetURLParams) (string, error) {
	if f.Config.ClientID == "" {
		skyErr := skyerr.NewError(skyerr.InvalidArgument, "ClientID is required")
		return "", skyErr
	}
	p := authURLParams{
		prividerName:   f.Config.Name,
		clientID:       f.Config.ClientID,
		urlPrefix:      f.Setting.URLPrefix,
		scope:          GetScope(params.Scope, f.Config.Scope),
		options:        params.Options,
		stateJWTSecret: f.Setting.StateJWTSecret,
		state:          NewState(params),
		baseURL:        f.BaseURL,
	}
	return authURL(p)
}
