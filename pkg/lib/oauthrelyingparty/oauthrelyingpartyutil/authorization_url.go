package oauthrelyingpartyutil

import (
	"net/url"
	"strings"
)

type AuthorizationURLParams struct {
	ClientID     string
	RedirectURI  string
	Scope        []string
	ResponseType string
	ResponseMode string
	State        string
	Prompt       []string
	Nonce        string

	ExtraQuery url.Values
}

func (p AuthorizationURLParams) Query() url.Values {
	v := url.Values{}

	v.Set("redirect_uri", p.RedirectURI)

	if p.ClientID != "" {
		v.Set("client_id", p.ClientID)
	}
	if len(p.Scope) > 0 {
		v.Set("scope", strings.Join(p.Scope, " "))
	}
	if p.ResponseType != "" {
		v.Set("response_type", string(p.ResponseType))
	}
	if p.ResponseMode != "" {
		v.Set("response_mode", string(p.ResponseMode))
	}
	if p.State != "" {
		v.Set("state", p.State)
	}
	if len(p.Prompt) > 0 {
		v.Set("prompt", strings.Join(p.Prompt, " "))
	}
	if p.Nonce != "" {
		v.Set("nonce", p.Nonce)
	}

	for key := range p.ExtraQuery {
		v[key] = p.ExtraQuery[key]
	}

	return v
}

func MakeAuthorizationURL(base string, query url.Values) string {
	return base + "?" + query.Encode()
}
