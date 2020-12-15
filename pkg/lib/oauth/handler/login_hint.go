package handler

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
)

type AnonymousIdentityProvider interface {
	ParseRequestUnverified(requestJWT string) (r *anonymous.Request, err error)
}

type LoginHintResolver struct {
	Anonymous AnonymousIdentityProvider
}

func (r *LoginHintResolver) ResolveLoginHint(loginHint string) (interface{}, error) {
	if !strings.HasPrefix(loginHint, "https://authgear.com/login_hint?") {
		return nil, nil
	}

	u, err := url.Parse(loginHint)
	if err != nil {
		return nil, err
	}
	query := u.Query()

	switch query.Get("type") {
	case "anonymous":
		jwt := query.Get("jwt")
		request, err := r.Anonymous.ParseRequestUnverified(jwt)
		if err != nil {
			return nil, err
		}

		return webapp.AnonymousRequest{
			JWT:     jwt,
			Request: request,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported login hint type: %s", query.Get("type"))
	}
}
