package handler

import (
	"errors"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type ProxyRedirectHandler struct {
	OAuthConfig *config.OAuthConfig
	HTTPOrigin  httputil.HTTPOrigin
	HTTPProto   httputil.HTTPProto
	AppDomains  config.AppDomains
}

func (h *ProxyRedirectHandler) Validate(redirectURIWithQuery string) (*oauth.WriteResponseOptions, error) {
	u, err := url.Parse(redirectURIWithQuery)
	if err != nil {
		return nil, errors.New("invalid redirect URI")
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
		return nil, errors.New("invalid redirect URI")
	}

	useHTTP200 := false
	isValid := false
	for _, c := range h.OAuthConfig.Clients {
		client := c

		err = validateRedirectURI(&client, h.HTTPProto, h.HTTPOrigin, h.AppDomains, []string{}, redirectURI)
		if err == nil {
			isValid = true

			if client.UseHTTP200() {
				useHTTP200 = true
			}
		}
	}

	if !isValid {
		return nil, errors.New("redirect URI is not allowed")
	}

	return &oauth.WriteResponseOptions{
		RedirectURI:  u,
		ResponseMode: "query",
		UseHTTP200:   useHTTP200,
		Response:     make(map[string]string),
	}, nil
}
