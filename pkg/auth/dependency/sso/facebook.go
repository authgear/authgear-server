package sso

import (
	"strings"

	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type FacebookImpl struct {
	Setting Setting
	Config  Config
}

func (f *FacebookImpl) GetAuthURL(params GetURLParams) (string, error) {
	if f.Config.ClientID == "" {
		skyErr := skyerr.NewError(skyerr.InvalidArgument, "ClientID is required")
		return "", skyErr
	}
	if params.UXMode == WebPopup {
		// https://developers.facebook.com/docs/facebook-login/manually-build-a-login-flow
		params.Options["display"] = "popup"
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

func (f *FacebookImpl) HandleAuthzResp(code string, scope Scope, encodedState string) (string, error) {
	h := authHandler{
		providerName:       f.Config.Name,
		clientID:           f.Config.ClientID,
		clientSecret:       f.Config.ClientSecret,
		urlPrefix:          f.Setting.URLPrefix,
		code:               code,
		scope:              scope,
		stateJWTSecret:     f.Setting.StateJWTSecret,
		encodedState:       encodedState,
		accessTokenURL:     AccessTokenURL(f.Config.Name),
		processAccessToken: f.processAccessToken,
	}
	return h.handle()
}

func (f *FacebookImpl) processAccessToken(a accessToken) accessToken {
	if a.ExpiresIn == 0 && a.RawExpires != 0 {
		a.ExpiresIn = a.RawExpires
	}
	if strings.ToLower(a.TokenType) == "bearer" {
		a.TokenType = "Bearer"
	}
	return a
}
