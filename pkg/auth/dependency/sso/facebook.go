package sso

import (
	"fmt"
	"net/url"
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
	encodedState, err := ToEncodedState(f.Setting.StateJWTSecret, params)
	if err != nil {
		return "", err
	}
	v := url.Values{}
	v.Set("response_type", "code")
	v.Add("client_id", f.Config.ClientID)
	v.Add("redirect_uri", RedirectURI(f.Setting.URLPrefix, f.Config.Name))
	v.Add("state", encodedState)
	v.Add("scope", strings.Join(params.Scope, " "))
	for k, o := range params.Options {
		v.Add(k, fmt.Sprintf("%v", o))
	}
	if params.UXMode == WebPopup {
		// https://developers.facebook.com/docs/facebook-login/manually-build-a-login-flow
		v.Add("display", "popup")
	}
	return BaseURL(f.Config.Name) + "?" + v.Encode(), nil
}
