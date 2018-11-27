package sso

import (
	"fmt"
	"net/url"
	"strings"
)

type authURLParams struct {
	prividerName   string
	clientID       string
	urlPrefix      string
	scope          Scope
	options        Options
	stateJWTSecret string
	state          State
	baseURL        string
}

func authURL(params authURLParams) (string, error) {
	encodedState, err := EncodeState(params.stateJWTSecret, params.state)
	if err != nil {
		return "", err
	}
	v := url.Values{}
	v.Set("response_type", "code")
	v.Add("client_id", params.clientID)
	v.Add("redirect_uri", RedirectURI(params.urlPrefix, params.prividerName))
	v.Add("state", encodedState)
	v.Add("scope", strings.Join(params.scope, " "))
	for k, o := range params.options {
		v.Add(k, fmt.Sprintf("%v", o))
	}
	return params.baseURL + "?" + v.Encode(), nil
}
