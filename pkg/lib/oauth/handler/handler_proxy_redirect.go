package handler

import (
	"errors"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type ProxyRedirectHandler struct {
	OAuthConfig *config.OAuthConfig
	HTTPOrigin  httputil.HTTPOrigin
	HTTPProto   httputil.HTTPProto
	AppDomains  config.AppDomains
}

func (h *ProxyRedirectHandler) Validate(redirectURIWithQuery string) error {
	u, err := url.Parse(redirectURIWithQuery)
	if err != nil {
		return errors.New("invalid redirect URI")
	}

	// Remove the query and fragment before validation
	redirectURI := &url.URL{
		Scheme: u.Scheme,
		Opaque: u.Opaque,
		User:   u.User,
		Host:   u.Host,
		Path:   u.Path,
	}

	if redirectURI.String() == "" {
		return errors.New("invalid redirect URI")
	}

	for _, c := range h.OAuthConfig.Clients {
		client := c
		err = validateRedirectURI(&client, h.HTTPProto, h.HTTPOrigin, h.AppDomains, []string{}, redirectURI)
		// pass the validation in one of the OAuth clients
		if err == nil {
			return nil
		}
	}

	return errors.New("redirect URI is not allowed")
}
